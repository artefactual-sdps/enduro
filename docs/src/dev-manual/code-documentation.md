# Code documentation

This guide describes how Enduro documents Go, JavaScript, TypeScript, and Vue
code. Its purpose is not to increase the number of comments, but to make each
comment intentional, useful, and consistent with the tools and conventions of
its language.

Code documentation has two audiences:

- callers need to understand the contract of a package, type, function,
  composable, or component; and
- maintainers need to understand the context behind an implementation.

These needs call for different kinds of comments.

## Philosophy

Code should explain its own mechanics through clear names, precise types, and
small, cohesive units. A comment should preserve information that the code
cannot express clearly by itself.

For implementation comments, document context and intent rather than
line-by-line mechanics. "Document why, not how" is a useful shorthand. Record
context such as:

- the reason for a design choice or an unusual sequence of operations;
- constraints, invariants, and assumptions;
- trade-offs and rejected alternatives when they still matter;
- lifecycle and cleanup responsibilities;
- side effects and concurrency constraints;
- compatibility workarounds and the conditions under which they can be removed;
  and
- behavior that is intentionally surprising or easy to break.

Do not translate a statement or a well-named function into prose. If a comment
only describes what the next line does, first consider whether clearer code
would make the comment unnecessary. Comments that divide a long orchestration
function into meaningful phases can still be useful, but they should aid
navigation rather than narrate each operation.

API documentation is different. It should describe the purpose of an
abstraction, how callers use it, and its observable behavior. Include special
results, errors, side effects, ordering, or concurrency guarantees when they are
part of the contract, but leave replaceable implementation details out.

[Google's code review guidance] makes the same distinction between comments
that explain why code exists and API documentation that explains purpose,
usage, and behavior. Matt Duck's [notes on A Philosophy of Software Design]
summarize the underlying principle: "Comments should be used to describe things
that aren't obvious from the code." The article also explores the different
purposes of interface and implementation comments.

Keep comments close to the code they describe and update or remove them when
the code changes. A stale comment is more harmful than a missing one. Do not
edit generated code to improve its comments; change the source or generator
instead.

## Go

Follow the official [Go doc comment] conventions:

- Give each package and each exported type, function, method, constant, and
  variable a doc comment.
- Start a package comment with `Package` followed by the package name.
- Start other doc comments with the declared name and write complete sentences.
- Describe what a function does or returns and the special cases callers need
  to understand.
- Document non-obvious zero-value behavior and concurrency guarantees on types.
- Keep algorithm and implementation details in comments inside the function,
  unless a property such as complexity or stability is part of the contract.

Go doc comments use `//` and are displayed by `go doc`, `pkg.go.dev`, and Go
language tooling. Let `gofmt` and the project's Go formatters format them.

For example:

```go
// NewFilter returns a new Filter. It panics if orderingFields is empty.
func NewFilter(query Query, orderingFields []string) *Filter {
    // ...
}
```

## JavaScript, TypeScript, and Vue

Use ordinary `//` comments by default. Enduro does not generate JavaScript or
TypeScript API documentation with JSDoc, and most exports do not need
documentation comments.

A shared API whose contract needs to be visible at call sites is a narrow
exception. TypeScript-aware editors attach `/** ... */` documentation comments
to symbols and display them at definitions, imports, and call sites. They do
not provide the same IDE documentation for ordinary `//` comments.

Use a `/** ... */` comment for a shared function, composable, class, or type
only when callers need to know something that its name and TypeScript signature
do not express. In addition to the general guidance above, useful details can
include:

- special return values, such as a value reserved for cancellation;
- errors callers can expect; and
- required integration with a Vue template or another framework.

Start with a short summary of the contract. Add paragraphs for important
semantics, and use tags such as `@example` and `@throws` when they make the API
easier to use. Avoid `@param` and `@returns` entries that merely repeat names
and types already present in the signature.

For example:

```ts
/**
 * Opens a component as a dialog and resolves with its result.
 *
 * Resolves with `undefined` if the host is removed before the dialog finishes.
 *
 * @throws `DialogAlreadyOpenError` when another dialog is already active.
 */
async function openDialog<Dialog extends DialogComponent>(
  component: Dialog,
  ...args: DialogArguments<DialogProps<Dialog>>
): Promise<DialogResult<Dialog> | undefined> {
  // ...
}
```

For Vue components, prefer explicit `defineProps` and `defineEmits` types,
descriptive slot and event names, and clear template structure. Add comments
only for contracts or lifecycle constraints that those declarations cannot
express. Use HTML comments in templates sparingly; they should explain
non-obvious structure rather than label self-explanatory markup.

This narrow `/** ... */` exception is not a reason to document every export.
Continue to use `//` for implementation context. Longer architecture,
multi-step workflows, and development or testing instructions belong in
project documentation rather than source-code comments.

[Go doc comment]: https://go.dev/doc/comment
[Google's code review guidance]: https://google.github.io/eng-practices/review/reviewer/looking-for.html
[notes on A Philosophy of Software Design]: https://www.mattduck.com/2021-04-a-philosophy-of-software-design.html
