# Development Best Practice

Software development is like nature: it's in a constant state of flux and continuously changing. As software engineers, it's our responsibility to embrace the flux and write clean code that is _optimized for change_. In doing so, we can rapidly innovate and satisfy customer needs.

## Guiding principles

These principles provide a guide for designing software. At times they may seem contradictory, but they ultimately aim to produce the same kind of code: _simple_ and _maintainable_.

> "Debugging is twice as hard as writing the code in the first place. Therefore, if you write the code as cleverly as possible, you are, by definition, not smart enough to debug it." - _Brian W. Kernighan_

- [SOLID](https://www.digitalocean.com/community/conceptual_articles/s-o-l-i-d-the-first-five-principles-of-object-oriented-design#single-responsibility-principle) - The Single-responsibility Principle, Open-closed Principle, Liskov Substitution Principle, Interface Segregation Principle and Dependency Inversion Principle collectively define a guide for writing maintainable code. While often referenced in the context of an object-oriented language, they are applicable to all languages.
- [LoD](https://en.wikipedia.org/wiki/Law_of_Demeter) - The Law of Demeter tells us a construct should talk to their direct dependencies and _only_ their direct dependencies (don't talk to strangers). Reaching to transitive dependencies creates complex layers of interaction that drive toward [spaghetti code](https://en.wikipedia.org/wiki/Spaghetti_code).
- [KISS](https://people.apache.org/~fhanik/kiss.html) - "Keep it simple, stupid" was coined by the US Navy. Systems should be designed as simply as possible.
- [YAGNI](https://martinfowler.com/bliki/Yagni.html) - "You aren't gonna need it" if you don't have a concrete use-case, don't write it.
- [DRY](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself) - "Don't repeat yourself" suggests you should preference code re-use over duplication. However, you should [avoid hasty abstractions](https://sandimetz.com/blog/2016/1/20/the-wrong-abstraction) as the wrong abstraction can be extremely costly to correct. Instead, lean into small amounts of duplication to help identify the right abstractions through multiple use cases and define the pathway to DRY design.

> "Leave code better than you found it"

Generally, you want to fix problems as they're identified. This includes micro refactors to accommodate new code that causes existing abstractions to, possibly, no longer make sense. Resist bolting on new code that is uncomplimentary to existing code. Refactoring efforts generally belong on their own PR thereby following the SRP guideline for PRs.

Finally, we have delivery commitments. Larger more intentional refactors may be required to fix a problem you've identified and attempting to tackle it may delay your deliverable. In such a case, write down your ideas and start a conversation with the wider team. 

## Defining _"clean code that is optimized for change"_

What constitutes clean code? We want simple and maintainable, but how do we define that? The following best practices offer niche advice. Collectively, it represent our definition of clean code.

### State

> Shared mutable state is believed by many to be the “root of all evil”, or at least the cause of most of the accidental complexity in our code. And “Complexity is the root cause of the vast majority of problems with software today.” - [Mauro Bieg](https://mb21.github.io/blog/2021/01/23/pure-functional-programming-and-shared-mutable-state.html#the-root-of-all-evil)

Package level state is rarely required. If it is, it should be composed of re-usable constructs defined and exposed from the package itself.

In avoiding package level state we avoid [hidden dependencies and unintended side effects](https://dave.cheney.net/practical-go/presentations/gophercon-israel.html#_avoid_package_level_state) stemming from global state mutation. Model loosely-coupled components by [declaring your dependencies](#dependencies).

### Names

##### Packages

Package names help form the first impression of what a package provides and is used to reference the public API of the package. A good package name brings about clarity and purpose for consumers. Package names should be concise and represent the behavior provided by the package, not the types it contains. 

Avoid stutter in struct and interface names by treating the package name as a namespace and considering how the types are consumed. For example, `drink.DrinkDecorator` should be `drink.Decorator`. You may find you have a type named after the package, sometimes it's appropriate and you should lean on the reviewer process to make a judgment call.

Prefer using the singular form of nouns for package names. If you find yourself using an appropriate package name that's commonly used as a variable, consider pluralizing the package name. For example, `string` is a type that could be shadowed hence the stdlib has a `strings` package.

Avoid [relentlessly] `types`, `interfaces`, `common`, `util` or `base` packages. These names offer no indication on what the package provides beyond 'a bundle of stuff'. With time, these types of package turn into dumping grounds containing unrelated logic that can be the root cause for cyclic dependencies.  If a utility style package is required, make it specific. For example a `/cmdutil` package may provide command utility functions.

##### Function & method

Functions and methods (functions with a receiver parameter) should adequately describe their behavior. Generally, they should follow a verb-noun form. The types of parameters should help inform the function name.

Capitalize acronyms in functions. When creating an unexported function starting with an acronym, lower-case the whole acronym.

##### Fields & variables

Variable names should be concise and descriptive. Prefer single word names. The further away from the site of declaration a variable is used, the more descriptive it needs to be. For example, a variable named `n` used 30 lines after declaration makes it unnecessarily difficult to reason what it represents at the site of use.

Avoid acronyms that aren't well known. For example, `kcp` should be `kubeadmControlPlane`. If in doubt, prefer clarity over brevity.

Capitalize acronyms in fields. When creating an unexported field starting with an acronym, lower-case the whole acronym.

### <a name="dependencies"></a> Dependencies

Functions should explicitly declare their dependencies as parameters. In doing so they make clear to consumers what is required and increase their flexibility and configurability.

```go
func FastenPlanks(tools Toolbox, surface Surface, planks Planks) error {
    for ; planks.Count() > 0;  {
        if err := tools.Hammer(surface, planks.Next()); err != nil {
            return err
        }
    }
}
```

Avoid constructing dependencies in-line. If you do construct a dependency in-line because it compliments the most common usage, provide a way to override it such as with a setter.

##### "Accept interfaces, return structs"

The Go community talks of the "accept interfaces, return structs" adage. The meaning of this idiom can be hard to pin down, herein lies some clarity.

"Accept interfaces" promotes flexible and loosely coupled code. When a function declares dependent behavior as an interface parameter it gives control of _how_ that behavior is satisfied to the consuming logic.

"return structs" is less obvious. It, generally, speaks to initializing constructors where authors should return the struct being constructed as opposed to interface it satisfies. This is largely due to duck typing (["If it walks like a duck and it quacks like a duck, then it must be a duck"](https://en.wikipedia.org/wiki/Duck_typing)).

> Interfaces in Go provide a way to specify the behavior of an object: if something can do this, then it can be used here. - [Effective Go](https://go.dev/doc/effective_go#interfaces_and_types)

Factory functions are not to be confused with initializing constructors. Factory functions typically return abstractions (an interface) satisfied by 1 of several implementations. A factory could be as simple as:

```go
// Foo and Bar satisfy fmt.Stringer
func NewStringer(key string) (fmt.Stringer, error) {
    switch key {
    case "foo":
        return Foo{}, nil
    case "bar":
        return Bar{}, nil
    default:
        return nil, errors.New("unknown key")
    }
}
```

Factories generally shouldn't contain complex construction logic. That's the job of the initializing constructor for the concrete type.

##### Loggers

Loggers are sometimes treated as a special dependency by making entire projects depend on a package level logger. This _masks_ a crucial dependency that is hard to change without impacting the rest of the code base. To understand why this is bad, consider extracting a package into its own repository for reusability. Depending on a package level logger, or any other package level dependencies, makes the extraction process unnecessarily difficult.

Logger's are not special dependencies, inject them like any other dependency. If you want to minimize the surface area of the default initializing constructor (`New...` funcs) you may construct noop loggers to satisfy the logger dependency and provide a means to override the logger with a setter.

### Testing

Software development is costly and hazardous. Tests provide example usage, guard rails, and assurances. Without tests, you cannot have high quality code. For this reason, all code we spend time designing and writing is worth testing. Similarly, if it can be unit tested, it should be unit tested. Unit tests are the first line of defense and can be complimented with additional integration or end-to-end tests.

Unit tests should:

- focus on the public API of a package.
- use the `_test` idiom for testing files.
- treat the package as a black box.
- cover at least 80% of the public API.
- named after the entity they're testing; for example, `TestFooBarDependencyErrors` tests that an error occurs when `Foo` is executed and its `Bar` dependency errors.

Testable code should:

- isolate or provide overrides such as setter functions for hard to test concerns like IO, time and concurrency.
- expose a flexible public API.

If you public API is hard to test consumers will find it hard to use - rethink your design.

##### `_test` idiom

The Go compiler allows test files, files suffixed with `_test.go`, to reside in an `_test` package. For example:

```go
// foo.go
package foo

// implementation
```

```go
// foo_test.go
package foo_test

// test cases
```

The compiler will build files that declare a package with the suffix "_test" as a separate package, then link and run with the main test binary. This ensures tests cases only have access to the public API of a package.

Test one scenario per test. This makes reading and debugging tests easier plus it keeps the test name simple (remember that the last part of the test function name describes the scenario you are testing).

Avoid having logic in tests. If you need to make your assertions conditional, that's a good indicator that the test should be split.

When testing the same scenario with multiple combinations of input/output, prefer table tests. This is not only common practice in `go` but it facilitates adding more cases when the tested code is changed.

##### Necessarily complex and unexported functions

If you have attempted to break a function down into singular responsibilities and found its best to maintain the necessary complexity as a single unexported function it may be appropriate to white box test.

When testing a _necessarily complex_ function isolate the tests to a separate testing file. For example `necessarily_complex_func_wb_test.go` where `wb` stands for white box.

### Errors

When generating an error, concisely describe the problem that occurred. Do not prefix the message with the word 'error' instead leaving that to a presentation layer of code. Use lower case unless describing a type the _user_ is aware of (generally a user is type and parameter unaware). 

When handling an error, don't log it only to pass it up the call stack. Logging and passing the received error up the call stack typically results in lots verbose and unhelpful logging about the same error.

If appropriate, add context by wrapping the error using `fmt.Errorf` or equivalent. Only add context that is known by the scope of your function and do not _assume_ an error represents something unless you can make a positive assertion. This avoids embedding knowledge that makes re-use and refactors harder. For example:

```go
// foo knows bar can return errZero so can perform an explicit check.
func foo() error {
    err := bar(0)
    if errors.Is(err, errZero) {
        return fmt.Errorf("ambiguous: %v", err)
    }

    if err != nil {
        return err
    }
}

var errZero = errors.New("zero value")

func bar(i int) error {
    if i > 5 {
        return errors.New("out of bounds")
    }

    if i == 0 {
        return errZero
    }
}
```

### Boolean parameters

Boolean parameters are widely considered an anti-pattern. Context is hugely important when assessing a boolean parameter but in general they represent an on/off switch for behavior. The behavioral on/off switch creates hard to understand and maintain logic that often has poor re-use and is an indicator that a piece of code violates the SRP. 

Instead, consider structuring your code to have dynamically constructible behavior logic. This allows you to isolate booleans to construction logic, a distinct concern, and construct the behavior you require for your context. 

An example of code designed to be dynamically constructed may include the [decorator](https://refactoring.guru/design-patterns/decorator) pattern. It may use a [builder](https://refactoring.guru/design-patterns/builder) or an [abstract factory](https://refactoring.guru/design-patterns/abstract-factory) with a builder to construct the logic using boolean flag configuration read from CLI flags. There are many more combinations and, generally, you want to consider your problem space and design a piece of code that makes sense for your problem first.

### Concurrency

Concurrency introduces harder to rationalize execution relative to serialized programs and is easy to get wrong. Concurrency doesn't necessarily speed up a program either. With these points in mind, prefer synchronous APIs instead leaving concurrency to the caller. Not only does this eliminate a whole host of problems, it ensures your APIs are predictable and easier to use.

Dave Cheney has made some excellent [comments](https://dave.cheney.net/practical-go/presentations/gophercon-israel.html#concurrency) on concurrency in Go.

#### Channels

If you find the need for channels prefer accepting them as parameters. Accepting channels as a parameter allows consumers to specify channel properties, such as buffered or unbuffered, creating a more flexible API. 

Clearly document channel ownership thereby documenting who is responsible for closing it. If there are multiple writers, no-one can close it therefore no-one owns it. Note channels don't _need_ to be closed, they will be garbage collected once the necessary scoping conditions are met.

Clearly document channel behavior. For example, what does your algorithm do if it can't send data on a channel? Will it drop data, block until the channel is writable, or something else?

Avoid using channels for data mutation operations where a traditional mutex would suffice. 

###  Comments

Functions and variables should be named so that their purpose is clear. However, good naming and well structured code isn't justification to not document APIs. All public package APIs should be documented including those that may seem trivial. This, at minimum, helps with code documentation generation. Additionally, each package should have a clear purpose that can be documented with examples using the [package documentation feature of Go](https://go.dev/doc/effective_go#commentary).

Through good documentation that we reflect on and update regularly, we can clearly communicate with maintainers what the original intent is making decisions easier when new capabilities are required.

### Returns

Return early instead of nesting deeply. The `else` statement is rarely required when following this rule. For example:

```go
func foo(bar string) error {
    if bar == "" {
        return errors.New("bar is empty")
    }

    // do something

    return nil
}
```

In returning early, we also create clearer code that has lower [cyclomatic complexity](https://en.wikipedia.org/wiki/Cyclomatic_complexity) that can be the cause for bugs.

Never use naked returns. Go allows return types to be specified with an identifier. The identifier is a scoped variable in the context of the function and its value can be set during program execution. Doing so, however, often creates unnecessarily difficult to rationalize logic.

Use named returns only if strictly necessary for disambiguation (for example, a function that returns the same type twice). If you don't specify names for your return values you can't accidentally use naked returns. 

### Panicking

In general, avoid panics. Panicking from a package implies the package understands the complete context of execution which is rarely the case. Instead, return errors wrapped with appropriate context describing the problem.

Herein lies an exception. When code is aware it is executing under a `main` func, or a derivative of `main` such as a `Command` object, it may have grounds for panicking if it identifies the program is invalid. Program invalidity is typically representative of a programmer error. For example, consider a `Command` object that requires, under certain circumstances, a `foo` flag to be configured as part of CLI argument parsing. If the program finds that flag doesn't exist (not to be confused with "wasn't set by the user"), the program is invalid. This is indicative of a programmer error as they _forgot_ to add the CLI flag. In this instance it is impossible for a user to fix the problem, because the program is invalid. Consequently, the program has grounds to panic. It is important to recognize that the panicking code is _fully aware_ of its execution context.

### `replace` directive

> A replace directive replaces the contents of a specific version of a module, or all versions of a module, with contents found elsewhere. The replacement may be specified with either another module path and version, or a platform-specific file path.

In general, avoid using `replace`.
These directives are not inherited by importing modules, making dependant modules have to replicate them and keep them in sync.
This pollutes the `go` ecosystem.

You need a good reason to use it (mostly if this is your last alternative) and you need an exit plan.
If you do, add a comment in `go.mod` explaining why that `replace` instance is needed and when (or under what conditions) it can be removed.

Some examples of situations where you might need a `replace`:
* Fixing CVE's in indirect dependencies. Make sure you specify the transitive dependency/dependencies so we can track when it gets updates upstream.

## Style

### Variable declaration

When declaring and not initializing, prefer `var`. For example, `var vehicle Vehicle`.

When declaring and initializing, use `:=`. For example, `vehicle := NewVehicle()`

Make intent clear for complicated variable initialization (contradictory to the 2 above rules). For example `var p uint32 = 0x80` makes a statement about the type as opposed to `p := uint32(0x80)` that focus' on the value, prefer the former.

### Line length

Traditionally, many coding standards have stipulated 80 chars as the maximum line length. It is thought to originate from the days of punch cards where IBM used 80 column widths. The width translated well to small width terminal monitors hence was adopted in the early days of computing.

In the present day we still see 80 chars line length feature in coding standards. However, much has changed since the standard was originally employed. Attempts, through research, to pin down the optimal line length exists but have resulted in conjecture and contradiction. For this reason, we consider other use-cases and constraints to help decide line length:

- laptops with less than 16" wide monitors are in abundance.
- side-by-side comparison of files is useful.
- horizontal scrolling is annoying.
- some developers _need_ larger font sizes.

When developing code, we ask you to take these points into consideration and not to create obnoxiously long lines. Compliment existing code. Configure your IDE to plot margins so it may aid your decision. A good rule of thumb is the 100 characters mark give or take 20.

### File naming

All `go` file names must follow `snake_case`. Using snake case compliments the Go compiler that expects test files to be appended with `_test`.

## References

These references are the basis for this document.

- Practical Go - Dave Cheney - https://dave.cheney.net/practical-go/presentations/gophercon-israel.html
- Best Practices - Peter Bourgon - https://peter.bourgon.org/go-best-practices-2016/
- Effective Go - The Go Team - https://go.dev/doc/effective_go
- Code Review Comments - The Go Community - https://github.com/golang/go/wiki/CodeReviewComments

## To do

Sections we want to write about with some raw notes.

**Package API**

- this might be wrapped up in naming and testing. Do we have more explicit points?

**Interfaces**

- Define the behavior a _consumer_ expects
- Premature interfaces
- Accept interfaces
- Should be cohesive
- Should be small
- Preferably single method

**Abstractions**

- Functions & Methods
- Constructs
- Interfaces

**Disambiguating context.Context**

- Represents the context of execution
    - Useful for tear down
- There should be 1 and only 1 context
- Should be the first parameter on the primary call path

**Logging**

What constitutes good logging? What should and shouldn't be logged?
