# `filewriter.FileWriter` refactor proposal

## Context

As an end user leverages the EKS-A tool various files are written to non-volitile storage. Some of these files are intended to remain in storage while others are transient and subseqently removed as the program executes. The `/pkg/filewriter` package provides a path context capability that ensures constructs within the EKS-A code perform io in a consistent location. The package contains a single interface and implementation. The interface defines methods for 3 capabilities: (1) defining nested path contexts; (2) cleaning up files within the instances path context by deleting the directory; and (3) writing files under the path context.

```go
type FileWriter interface {
	Write(fileName string, content []byte, f ...FileOptionsFunc) (path string, err error)
	WithDir(dir string) (FileWriter, error)
	CleanUp()
	CleanUpTemp()
	Dir() string
}
```

This proposal seeks to both evaluate the existing use of the `FileWriter` interface and use that data to inform a refactor that increases code flexibility, promote better ownership of files and enable a stronger separation of concerns.

## Problem

The `FileWriter` interface exposed by the `filewriter` package has multiple concerns: (1) ensuring the path context a given instance represents exists; (2) writing data to a file name under the path context and setting the permissions of the file; (3) cleaning up the path context; (4) managing temporary files located in a sub directory under the path context. If the `CleanUp()` method is invoked it also cleans up ay temporary files. With respect to temporary files, any call to the `Write(...)` method is, by default, assumed to be temporary and written in a special location under the path context (`<context root>/generated`).

While the interface helps solve the problem of ensuring consuming constructs write files to a shared context while removing the need to know the context, it also forces consumers to absorb additional concerns that make it difficult to apply the SRP. Specifically, the `Write(...)` method requires consumers to know what filename they want to use and define any available options for that file.

When writing files, one may intutively think they're written under the path context and considered persistent but `Write(...)` has confusing semantics that mean files are written as temporary under a sub directory. Of the 26 calls to `Write(...)` that I found and evaluated, 19 of them declared the file persistent using the `PersistentFile` option. Some of the remaining calls were incorrectly using a `FileWriter` interface and others wrote files but took no ownership of cleaning them up. This suggests in the majority of cases consumers want to write persistent files and in some instances the semantics of `Write()` were perhaps unclear.

On ownership, the `FileWriter` interface doesn't make clear who is responsible for cleaning up files resulting in ambiguity in consuming code. When considering the `CleanUp()` and `CleanUpTemp()` calls in conjunction with `WithDir()` its difficult to know whether its safe for a top level `FileWriter` to clean up given there's no way to distinguish permanent and temporary files in path contexts created with `WithDir()` contributing further ambiguity to ownership of file lifecycle. When evaluating calls to `CleanUp()` we found it was not called anywhere in the code. With respect to `CleanUpTemp()` we found only a single functional use at the root file writer instance where temporary files are cleaned only if the callpath does not experience an error. Some calls to `CleanUpTemp()` were made immediately after creating a new `FileWriter` instance suggesting the author understood the `NewWriter()` semantics and wanted to immediately delete the `generated` directory `NewWriter()` creates for temporary files.

With respect to the `FileWriter` interface composition, `FileWriter` defines all methods for the concerns outlined earlier directly. This forces consumers to depend on the full `FileWriter` interface even if they do not need to define new path contexts. The vast majority of consumers only care about writing files hence the other 4 methods are superfluous in most contexts.

Finally, the `Write()` method writes files with unix permission bits 0777 before the umask is applied. This brings about risk as files are written with the execution bit for both group and other assuming a default umask of 022.

## Proposal

I propose we reduce the scope of the `FileWriter` to simply providing convinience methods for writing under a path context. Reducing the scope includes removal of the `CleanUp()` and `CleanUpTemp()` methods and any persistent or temporary file options. Additionally, we make clear consumers are responsible for cleaning any files they create with code docs on creation methods. If consumers wish to delete the whole directory they can use the standard libraries `os.RemoveAll(w.Dir())`. Additionally, any files created are assumed to be permanent. Consumers can explicitly define temporary file path contexts that are managed as independent contexts if they need to model temporary files. Finally, I propose a breakdown of interfaces to allow consumers to depend on only what they need and an introduction of a method set that compliments the standard libraries `io` package. Complienting the standard libraries `io` package ensures consumers can better satisfy the SRP as they can function without needing to manage file properties such as name and permissions.

With respect to the default permissions assigned to files, I propose we change to using 0644 granting read and write to the owner and read permissions to group and other.

## Implementation

The following interfaces accept names as at least 1 parameter. The default implementation will prepend the path context to the name and return the `*os.File` type that can be used in `io` interfaces such as `io.Reader` and `io.Writer`.

```go
// Opener has the same behavior as os.Open().
type Opener interface {
    Open(name string) (*os.File, error)
}

// OpenFiler has the same behavior as os.OpenFile().
type OpenFiler interface {
    OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
}

// Creator has the same behavior as os.Create().
type Creator interface {
    Create(name string) (*os.File, error)
}

// WriteFiler performs the same function as the existing Write() call but mimics os.WriteFile() with the addition
// of the path as a return value.
type WriteFiler interface {
    WriteFile(name string, data []byte, perm os.FileMode) (n int, path string, err error)
}

type FileWriter interface {
    Opener
    OpenFiler
    Creator
    WriteFiler

    // Dir retrieves an absolute path to the current path.
    Dir() string

    // New creates a new FileWriter with subdir appended to the current path. Its the callers responsibility to delete
    // the context if desired.
    New(subdir string) FileWriter
}
```

Most consumers will only want an `Opener` or a `Creator` but the other interfaces exist to provide greater flexibility when customization of file properties is required. By using `Opener` and `Creator` we inherit the default file permission of 0644 and enable users to leverage functions from the standard library they're more likely to be familiar with.

An additional interface we can consider reads from an `io.Reader` when writing new files increasing the flexibility further.

```go
// WriteFileFromer works much the same as WriteFile but reads from an io.Reader better complimenting
// the io package.
type WriteFileFromer interface {
    WriteFileFrom(name string, r io.Reader, perm os.FileMode) (n int, path string, err error)
}
```

## Transition

Given most calls to `Write()` declare them permanent the temporary file implementation can likely be removed trivially. First, we can modify the behavior of `Write` such that all files are considered permanent. This has the effect of leaving the few instances of temporary file writes under the root path's `generated` dir on disk which should be harmless to the end user. Having modified the default `Write()` behavior we can strip the code base of the `PermanentFile` option normalizing `Write()` calls. While this effectively renders current temporary files as permanent, given the limited set of uses for writing temporary files we can modify the code to ensure it takes ownership for deleting the temporary files it writes.

Second, we declare `WithDir()` deprecated as it results in the creation of the `generated` directory. We provide the `New()` alternative and begin updating calls to `WithDir()` with `New()`. Given all files are now written as permanent.

Third, we update all calls to `Write()` with `WriteFile()` to compliment its new signature. The behavior is largely the same.

Finally, we can update consumers to depend on the interface they require. For most consumers this will be the `WriteFiler` interface. As the code evolves we may create meta interfaces containing both `WriteFiler` and `Creator` so code can transition away from `WriteFiler` that still has the problem of forcing algorithms to deal with filenames.