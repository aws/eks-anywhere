# Development Best Practice

Software development is like nature: it's in a constant state of flux and continuously changing. As software engineers, it's our responsibility to embrace the flux and write clean code that is _optimized for change_. In doing so, we can rapidly innovate and satisfy customer needs.

## Contents

- [Development Best Practice](#development-best-practice)
  - [Contents](#contents)
  - [Guiding principles](#guiding-principles)
  - [Defining _"clean code that is optimized for change"_](#defining-_clean-code-that-is-optimized-for-change_)
  - [Additions](#additions)
    - [Global state](#global-state)
    - ["Accept interfaces, return structs"](#accept-interfaces-return-structs)
    - [Unit tests](#unit-tests)
    - [Avoid appending error messages with the word 'error'](#avoid-appending-error-messages-with-the-word-error)
    - [Avoid boolean parameters](#avoid-boolean-parameters)
    - [Comment package APIs](#comment-package-apis)
    - [Don't use naked returns](#dont-use-naked-returns)
    - [Panicking](#panicking)
    - [`replace` directive](#replace-directive)
    - [Line length](#line-length)
    - [File naming](#file-naming)

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

What constitutes clean code? We want simple and maintainable, but how do we define that? The Go community has developed a set of resources that can be considered part of our clean code definition. Take the time to _read them_ diligently and use them as references in code reviews when providing feedback.

|Title|URL|Description|
|---|---|---|
|Effective Go|https://go.dev/doc/effective_go|Official Go documentation on writing Go code|
|Code Review Comments|https://github.com/golang/go/wiki/CodeReviewComments|Community driven additions to Effective Go that are battle tested|
|Practical Go|https://dave.cheney.net/practical-go/presentations/gophercon-israel.html|Dave Cheney is a respected member in the Go community that has established further guidance based on experience|
|Style Guide|https://google.github.io/styleguide/go/decisions|Google developed a style guide that is referenced by the Code Review Comments|

The [Additions](#additions) section provides additional guidance for the EKS Anywhere code base that isn't covered by these articles or proves contradictory.

## Additions

### Global state

> Shared mutable state is believed by many to be the “root of all evil”, or at least the cause of most of the accidental complexity in our code. And “Complexity is the root cause of the vast majority of problems with software today.” - [Mauro Bieg](https://mb21.github.io/blog/2021/01/23/pure-functional-programming-and-shared-mutable-state.html#the-root-of-all-evil)

Package level state is rarely required. If it is, it should be composed of re-usable constructs defined and exposed from the package itself.

See https://dave.cheney.net/practical-go/presentations/gophercon-israel.html#_avoid_package_level_state).

### "Accept interfaces, return structs"

This section is included because the phrase surfaces regularly.

Code Review Comments provides a simple explanation that covers the general case. However, we want to stress this is the general case, not all cases.

When developing abstractions it may become necessary to return interfaces but this should be considered orthogonal to accepting interfaces.

See https://github.com/golang/go/wiki/CodeReviewComments#interfaces

### Unit tests

Unit tests should:

- focus on the public API of a package.
- prefer the `_test` when testing the public API of the package.
- treat the package as a black box.

Test names:
- When testing a function, the test name should start with the name of that function.
- When testing a method, the test name should start with the name of the receiver type followed by the method name.

When testing the same name using different test funcs append a concise test description to the test names.

If you public API is hard to test, consumers may find it hard to use - consider restructuring your package API.

### Avoid prepending error messages with the word 'error'

Prefixing error messages with the word error is typically unnecessary leading to error bloat when adding context; leave prefixing to presentation logic.

Also see https://github.com/golang/go/wiki/CodeReviewComments#error-strings.

### Avoid boolean parameters

We consider boolean parameters an anti-pattern. They indicate toggling behavior within an algorithm suggesting the algorithm has side effects and it makes rationalize behavior at the site of consumption difficult.

For example

```go
parsed := p.Parse(input, false) // What does false mean? What would happen if it were true?
```

###  Comment package APIs

At minimum, exported APIs should be documented. Consider `go doc` output when documenting APIs. Consumers of the APIs can see the function signature, not the implementation. Focus comments on what the function provides as opposed to the nitty gritty details of the implementation.

See https://github.com/golang/go/wiki/CodeReviewComments#doc-comments and https://github.com/golang/go/wiki/CodeReviewComments#package-comments.

### Don't use naked returns

Naked returns are unnecessary, just be explicit.

See https://github.com/golang/go/wiki/CodeReviewComments#naked-returns.

### Panicking

The need to panic only arises when the program enters an irrecoverable state. All other cases, especially when writing self contained packages, should return their errors.

See https://github.com/golang/go/wiki/CodeReviewComments#dont-panic.

### `replace` directive

> A replace directive replaces the contents of a specific version of a module, or all versions of a module, with contents found elsewhere. The replacement may be specified with either another module path and version, or a platform-specific file path.

Avoid using `replace` directives for long periods of time. If a `replace` directive is required, an exit plan to remove the replace is also required.

### Line length

Traditionally, many coding standards have stipulated 80 chars as the maximum line length. It is thought to originate from the days of punch cards where IBM used 80 column widths. The width translated well to small width terminal monitors hence was adopted in the early days of computing.

In the present day we still see 80 chars line length feature in coding standards. However, much has changed since the standard was originally employed. Attempts, through research, to pin down the optimal line length exists but have resulted in conjecture and contradiction. For this reason, we consider other use-cases and constraints to help decide line length:

- laptops with less than 16" wide monitors are in abundance.
- side-by-side comparison of files is useful.
- horizontal scrolling is annoying.
- some developers _need_ larger font sizes.

When developing code, we ask you to take these points into consideration and not to create obnoxiously long lines. Compliment existing code. Configure your IDE to plot margins so it may aid your decision. A good rule of thumb is the 100 characters mark give or take 20.

See https://github.com/golang/go/wiki/CodeReviewComments#line-length.

### File naming

All Go source files should be named with `snake_case` including files that represent types.
