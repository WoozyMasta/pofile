<!-- Automatically generated file, do not modify! -->

# Lint Rules Registry

This document contains the current registry of lint rules.

Total rules: 16.

## pofile

PO File

> Lint rules for Gettext PO parsing and semantic checks.

Rule groups for `pofile`:

* [lint](#lint)
* [parse](#parse)

### lint

> PO semantic lint diagnostics.

Codes:
[PO2001](#po2001),
[PO2002](#po2002),
[PO2003](#po2003),
[PO2004](#po2004),
[PO2005](#po2005),
[PO2006](#po2006),
[PO2007](#po2007),
[PO2008](#po2008),
[PO2009](#po2009),
[PO2010](#po2010),

#### `PO2001`

Duplicate `domain/msgctxt/msgid` entry key

> More than one entry has same `domain` + `msgctxt` + `msgid` key. Keep single
> canonical entry per key.

| Field | Value |
| --- | --- |
| Rule ID | `pofile.lint.duplicate-domain-msgctxt-msgid-entry-key` |
| Scope | `lint` |
| Severity | `error` |
| Enabled | `true` (implicit) |

#### `PO2002`

Entry has empty `msgid` value

> Entry is present but source key is empty. Remove broken entry or fill valid
> `msgid` value.

| Field | Value |
| --- | --- |
| Rule ID | `pofile.lint.entry-has-empty-msgid-value` |
| Scope | `lint` |
| Severity | `error` |
| Enabled | `true` (implicit) |

#### `PO2003`

`Msgstr[n]` present but `msgid_plural` is missing

> Plural translations require plural source form. Add `msgid_plural` when
> `msgstr[1]` or higher indexes are present.

| Field | Value |
| --- | --- |
| Rule ID | `pofile.lint.msgstr-n-present-but-msgid-plural-is-missing` |
| Scope | `lint` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `PO2004`

`Msgid_plural` present but plural `msgstr[n]` is incomplete

> Plural source form exists but translated plural slots are incomplete. Add at
> least `msgstr[1]` and other required indexes for locale rules.

| Field | Value |
| --- | --- |
| Rule ID | `pofile.lint.msgid-plural-present-but-plural-msgstr-n-is-incomplete` |
| Scope | `lint` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `PO2005`

Plural indexes have gap

> Plural entry has missing `msgstr[n]` between existing indexes. Use
> continuous index range without holes.

| Field | Value |
| --- | --- |
| Rule ID | `pofile.lint.plural-indexes-have-gap` |
| Scope | `lint` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `PO2006`

`Msgstr[n]` count mismatches `Plural-Forms`

> `Plural-Forms` header declares `nplurals`, but entry translations do not
> match expected plural slot count.

| Field | Value |
| --- | --- |
| Rule ID | `pofile.lint.msgstr-n-count-mismatches-plural-forms` |
| Scope | `lint` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `PO2007`

Duplicate header key

> Header contains same key more than once. Keep one canonical key/value entry
> to avoid parser-dependent behavior.

| Field | Value |
| --- | --- |
| Rule ID | `pofile.lint.duplicate-header-key` |
| Scope | `lint` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `PO2008`

Translations exist but `Language` header is empty

> File contains translated strings but header does not declare target
> language.

| Field | Value |
| --- | --- |
| Rule ID | `pofile.lint.translations-exist-but-language-header-is-empty` |
| Scope | `lint` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `PO2009`

Printf placeholders mismatch

> Set of printf-style verbs differs between source and translated string.
> Runtime formatting may fail or substitute wrong arguments.

| Field | Value |
| --- | --- |
| Rule ID | `pofile.lint.printf-placeholders-mismatch` |
| Scope | `lint` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

#### `PO2010`

Entry has empty translation text

> Translated entry contains empty `msgstr` value. Enable this check only when
> empty translations should be treated as potential issues.

| Field | Value |
| --- | --- |
| Rule ID | `pofile.lint.entry-has-empty-translation-text` |
| Scope | `lint` |
| Severity | `warning` |
| Enabled | `true` (implicit) |

### parse

> PO parser diagnostics.

Codes:
[PO1001](#po1001),
[PO1002](#po1002),
[PO1003](#po1003),
[PO1004](#po1004),
[PO1005](#po1005),
[PO1006](#po1006),

#### `PO1001`

Directive value must be quoted

> PO directives such as `msgid`, `msgstr`, `msgctxt`, and `msgid_plural` must
> use quoted string value on the same logical line.

| Field | Value |
| --- | --- |
| Rule ID | `pofile.parse.directive-value-must-be-quoted` |
| Scope | `parse` |
| Severity | `error` |
| Enabled | `true` (implicit) |

#### `PO1002`

Unknown PO directive or malformed line

> Parser found non-empty line that is not recognized as PO keyword, comment,
> or valid continuation content.

| Field | Value |
| --- | --- |
| Rule ID | `pofile.parse.unknown-po-directive-or-malformed-line` |
| Scope | `parse` |
| Severity | `error` |
| Enabled | `true` (implicit) |

#### `PO1003`

Continuation string is outside active field

> String continuation line must follow active field (`msgid`, `msgstr`,
> `msgctxt`, or `msgid_plural`) and cannot appear standalone.

| Field | Value |
| --- | --- |
| Rule ID | `pofile.parse.continuation-string-is-outside-active-field` |
| Scope | `parse` |
| Severity | `error` |
| Enabled | `true` (implicit) |

#### `PO1004`

`Header` metadata line is malformed

> `header` metadata line must follow supported syntax with valid key/value
> shape for this parser.

| Field | Value |
| --- | --- |
| Rule ID | `pofile.parse.header-metadata-line-is-malformed` |
| Scope | `parse` |
| Severity | `error` |
| Enabled | `true` (implicit) |

#### `PO1005`

`Msgstr[n]` index must be non-negative integer

> Plural translation form must use integer index notation `msgstr[n]` with
> non-negative numeric `n`.

| Field | Value |
| --- | --- |
| Rule ID | `pofile.parse.msgstr-n-index-must-be-non-negative-integer` |
| Scope | `parse` |
| Severity | `error` |
| Enabled | `true` (implicit) |

#### `PO1006`

Entry is missing required `msgid`

> Every non-obsolete PO entry must contain source identifier `msgid`.

| Field | Value |
| --- | --- |
| Rule ID | `pofile.parse.entry-is-missing-required-msgid` |
| Scope | `parse` |
| Severity | `error` |
| Enabled | `true` (implicit) |

---

> Generated with
> [lintkit](https://github.com/woozymasta/lintkit)
> version `dev`
> commit `unknown`

<!-- Automatically generated file, do not modify! -->
