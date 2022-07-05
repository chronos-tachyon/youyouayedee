# youyouayedee
UUID library for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/chronos-tachyon/youyouayedee.svg)](https://pkg.go.dev/github.com/chronos-tachyon/youyouayedee)
[![BSD 2-Clause License](https://img.shields.io/github/license/chronos-tachyon/youyouayedee)](https://github.com/chronos-tachyon/youyouayedee/blob/main/LICENSE)

Package youyouayedee provides tools for working with Universally Unique
Identifiers (UUIDs).

> A note for non-native English speakers and others confused by the package
> name: it is a phonetic spelling of the acronym "UUID".  Since Go package
> names get imported into the caller's namespace, I didn't want to use the
> name "uuid" because that's a very common variable name for UUID values.)

There are 5 well-known UUID versions defined in [RFC 4122][]:

Version 1 UUIDs are based on the current time (measured as hectonanoseconds
since 1582-10-12T00:00:00Z on the Gregorian calendar, *including* leap
seconds), a 14-bit rollover counter that is private to the generating host but
shared across all software on the host that generates V1 UUIDs, and (last but
not least) the generating host's "node identifier", which is traditionally the
MAC address of the host's network card but which may be something else for
privacy reasons, so long as it is stable and globally unique.  The fields are
also weirdly out of order, so they don't sort very well despite being
monotonic-ish.

> NB: V1 UUIDs are extremely popular despite the fact that absolutely nobody
> correctly follows the spec when generating them.

Version 2 UUIDs are based on the Open Software Foundation's Distributed
Computing Environment specification.  They are extremely rare.

Version 3 UUIDs are based on the MD5 hash of a namespace UUID and a string.
Mostly obsolete because of MD5.  Compare to V5 UUIDs.

Version 4 UUIDs are based on 122 bits chosen at random, plus 6 well-known bits
to make it a valid UUID.

Version 5 UUIDs are just like V3 UUIDs, except that the hash function is SHA1
instead of MD5.  Since malicious collision resistance is not actually a
significant use case for most users of UUIDs, there's nothing actually wrong
with V5 UUIDs despite SHA1 being extremely cryptographically broken in 2022.
However, neither V3 nor V5 UUIDs have any significant advantages over V4
UUIDs.  This library supports them but they are not recommended for general
use.  They can be good for generating well-known UUIDs defined in a
specification, however.

In addition, this library supports the 3 additional UUID versions defined in
the IETF document [draft-peabody-dispatch-new-uuid-format-04][], which this
library's author is very excited about.

Version 6 UUIDs are based on the current time, counter, and node identifier
just like V1 UUIDs, but their fields are rearranged to make them more sortable
in databases.  If you already have V1 UUIDs, you can convert them to V6 for
database storage and then back again if you need the exact same V1 UUID.

Version 7 UUIDs, which are meant to fully replace V1 and V6 UUIDs, are also
based on the current time.  Unlike V1 and V6, they are based on the well-known
and comparatively well-loved Unix `time_t` epoch (milliseconds since
1970-01-01T00:00:00Z, *excluding* leap seconds) and their meaning is thus much
easier to grok with the tools given to you by the OS.  The additional non-time
bits are now left as the implementor's choice, with random bits or monotonic
counters as proposed methods.  Also, none of the bits are dedicated to
sub-millisecond time precision, which few hosts are truly capable of providing
*anyway* because they are not equipped with locally installed atomic clocks
and NTP alone cannot achieve such accuracy.

Version 8 UUIDs are fully opaque, with their meaning defined exclusively by
the implementor.  As such, they cannot be expected to be "universally" unique
across all software and all machines, but they may be useful in specific
contexts.  This library has limited support for them, but you will need to
roll your own UUID generation algorithm (for obvious reasons).

[RFC 4122]: https://rfc-editor.org/rfc/rfc4122.html
[draft-peabody-dispatch-new-uuid-format-04]: https://datatracker.ietf.org/doc/html/draft-peabody-dispatch-new-uuid-format-04
