# Memory & Lifetime Audit — Phase 6 Plan 02

**Generated:** 2026-04-11
**Scope:** `src/fields2cover/`, `include/fields2cover/`, `swig/`
**Method:** Manual review (static tools cannot see ownership semantics reliably).
**Companion:** plan 06-01 static-analysis findings use F-NNN IDs; this doc
uses M-NNN IDs. Feeds plan 06-04 (low-risk fixes) and 06-05 (high-risk triage).

## Method

Manual review of:

1. Raw pointer declarations and manual `new` / `delete` expressions.
2. Rule-of-5 compliance on large types (`Cells`, `Swaths`, `Path`, `Field`, `Route`, `Swath`, `SwathsByCells`).
3. Pass-by-value sites for large types in public APIs.
4. SWIG binding boundary object lifetimes (`swig/Fields2Cover.i`).
5. External C-handle management (GEOS, GDAL, OGR).
6. Reference-return semantics on public accessors.

Commands used (via the Grep tool, not bash grep):

- `\bnew\b[^a-zA-Z_]` — manual allocations
- `\bdelete\b` — manual deallocations
- `reinterpret_cast|const_cast` — red-flag casts
- `return\s+&` — returning address of local/member
- `operator\s*=\s*\(` — custom assignment operators
- `~\w+\s*\(\s*\)` — destructors
- `GEOSGeom|OGRGeometry|GDALDataset|OGRSpatialReference` — C-library handles
- `shared_ptr|unique_ptr|weak_ptr` — smart-pointer usage sites
- `%newobject|%typemap|%ignore|%shared_ptr` — SWIG lifetime directives

Raw output lives in `build/audit/memory-survey.txt` (scratch, gitignored).
Intermediate review notes in `build/audit/memory-review.md` (scratch, gitignored).

## Summary

| Severity | Count |
|---|---|
| Critical | 1 |
| High     | 3 |
| Medium   | 5 |
| Low      | 3 |
| **Total**| **12** |

Key takeaways:

- Only **one** manual `new` and **one** manual `delete` exist in the entire
  library. OGR handles are otherwise correctly wrapped in `std::shared_ptr` or
  `std::unique_ptr` with custom deleters.
- `reinterpret_cast` and `const_cast` are used **zero** times. Excellent.
- `return &local` / `return &member` patterns: **zero** occurrences.
- The highest-value findings are around **rule-of-5 compliance** on
  collection types (`Swaths`, `SwathsByCells`) and the **`EmptyDestructor`
  non-owning wrapper hazard** in `Geometry<T,R>`.
- SWIG boundary is **safe by default** (copy-out on reference-returns) but
  that safety is invisible — a future `%typemap(out)` or `%naturalvar`
  addition could silently break it.

## Findings

### Critical

#### M-001 — `include/fields2cover/types/Geometry_impl.hpp:33,40` + `src/fields2cover/types/{Cell.cpp:42,53, Cells.cpp:68,76, MultiLineString.cpp:76,84, MultiPoint.cpp:39,47}`

**What:** `Geometry(T* g, EmptyDestructor)` and `Geometry(OGRGeometry* g, EmptyDestructor)` construct a `std::shared_ptr<T>` with a no-op deleter, producing a "view" wrapper whose underlying OGR pointer is owned by some *other* object. `Cells::getGeometry(size_t, Cell&)` and its siblings use this path to return borrowed references-disguised-as-values.

**Why it matters:** The returned `Cell` / `LinearRing` / `LineString` / `Point` has the same type as an owning wrapper. Nothing at the type level tells a caller that writing
```cpp
Cells cells = ...;
Cell c;
cells.getGeometry(0, c);
cells = Cells();   // parent dies, shared_ptr no-op deleter does nothing
c.area();          // UB: c's shared_ptr points to freed OGR memory
```
is a use-after-free. No compile-time or run-time warning. The `EmptyDestructor` tag is hidden from SWIG (`%ignore f2c::types::EmptyDestructor;` in `Fields2Cover.i:35`), so Python users cannot trip this directly — but any C++ caller can.

**Proposed fix:** one of
  (a) make `getGeometry(size_t, Cell&)` deep-copy (match the other overload `Cell Cells::getGeometry(size_t)` at `Cells.cpp:79` which already returns a clone);
  (b) introduce a distinct `CellView` / `CellRef` type so the non-ownership is visible in the type system;
  (c) rename `EmptyDestructor` to `NonOwning` and add `// NONOWNING VIEW — do not outlive parent` at every callsite.

**Estimated effort:** M — touches 6 accessor methods in `types/` plus tests in `tests/cpp/types/`.

### High

#### M-002 — `include/fields2cover/types/Swaths.h:25` + `src/fields2cover/types/Swaths.cpp:27`

**What:** `Swaths` has a user-declared `~Swaths()` but declares **no** copy ctor, move ctor, copy assign, or move assign. Per C++ rules, move operations are not implicitly generated when a destructor is user-declared. Copy operations are still implicitly generated (but deprecated).

**Why it matters:** Every `std::move(swaths)` and every `std::vector<Swaths>` reallocation silently copies instead of moves. `Swaths` holds `std::vector<Swath>`; copy-constructing touches every `Swath`, each of which holds a `LineString` with a `shared_ptr` to an OGR object. The copy increments atomic refcounts linearly. `Route` holds `std::vector<Swaths>` (Route.h:65) — so `Route::addSwaths()` / `addConnectedSwaths()` trigger reallocation-cascade copies on large fields. This is hot-path performance loss, not correctness.

**Proposed fix:** prefer rule-of-zero — drop `~Swaths()` entirely (there is nothing non-trivial to destroy). If a user-declared dtor is needed for ABI stability, add all five explicitly:
```cpp
Swaths(const Swaths&) = default;
Swaths(Swaths&&) noexcept = default;
Swaths& operator=(const Swaths&) = default;
Swaths& operator=(Swaths&&) noexcept = default;
~Swaths() = default;
```

**Estimated effort:** S.

#### M-003 — `include/fields2cover/types/SwathsByCells.h:22` + `src/fields2cover/types/SwathsByCells.cpp:28`

**What:** Identical pattern to M-002. User-declared `~SwathsByCells()`, no explicit copy/move. Member is `std::vector<Swaths>`.

**Why it matters:** Same reallocation-cascade cost, one layer deeper: copying `SwathsByCells` copies every `Swaths`, which copies every `Swath`.

**Proposed fix:** same — rule-of-zero or rule-of-five with `noexcept`.

**Estimated effort:** S.

#### M-004 — `include/fields2cover/types/Geometry_impl.hpp:329`

**What:** `delete poOGRProduct;` is called on a raw `OGRGeometry*` obtained from `OGRGeometryFactory::createFromGEOS` (via `buildGeometryFromGEOS` a few lines earlier). OGR objects created by the factory must be released via `OGRGeometryFactory::destroyGeometry()` to match the allocation family.

**Why it matters:** Mixing `new`/`delete` with factory `create`/`destroy` is undefined behavior when GDAL is built with a non-default allocator (some distro builds use `CPLMalloc`). The code may work on a standard GDAL build but is a latent ticking clock. It is also the *only* raw `delete` in the entire library — fixing it makes the library `delete`-free.

**Proposed fix:** replace with `OGRGeometryFactory::destroyGeometry(poOGRProduct);`. One-line change.

**Estimated effort:** S.

### Medium

#### M-005 — `include/fields2cover/types/Swath.h:23-33` + `src/fields2cover/types/Swath.cpp:23-26`

**What:** `Swath` declares `virtual ~Swath()`, `virtual operator=(Swath&&)`, `virtual operator=(const Swath&)`, and a non-virtual copy constructor, but **does not declare a move constructor**. With a user-declared copy ctor, the compiler does not generate a move ctor. Also, `operator=` is marked `virtual` on a class that has no subclasses in the library — this is a slicing hazard if anyone ever derives from `Swath`.

**Why it matters:**
- Moving a `Swath` (e.g. `std::vector<Swath>::push_back` during reallocation in `Swaths::emplace_back`) silently falls back to copy. `LineString` copy involves an atomic increment on the `shared_ptr` plus a structure copy.
- Virtual assignment is almost always wrong — derived-class state is not covered, and calling `base = derived` via `operator=` through a base pointer slices. No derived class exists today, but the `virtual` keyword invites one.

**Proposed fix:**
- Add `Swath(Swath&&) noexcept = default;` in Swath.h and `Swath::Swath(Swath&&) noexcept = default;` in Swath.cpp.
- Remove `virtual` from `operator=` unless polymorphic assignment is actually needed.
- Mark move operations `noexcept`.

**Estimated effort:** S.

#### M-006 — `include/fields2cover/types/Field.h:25,26`

**What:** `Field::Field(Field&&)` and `Field::operator=(Field&&)` are declared without `noexcept`.

**Why it matters:** `F2CFields = std::vector<F2CField>` (types.h:58) triggers the noexcept gate during reallocation. `std::vector` requires `is_nothrow_move_constructible_v<T>` to use the move path in `reserve()` / `push_back()` reallocation; otherwise it falls back to copy for the strong-exception guarantee. Field's members are all noexcept-movable (`std::string`, `Point`, `Cells`) so adding `noexcept` costs nothing and unlocks move semantics.

**Proposed fix:** add `noexcept` to the declaration and definition of Field's move ctor and move assign.

**Estimated effort:** S.

#### M-007 — `include/fields2cover/types/Robot.h:32,34`

**What:** `Robot`'s move constructor and move assignment are declared without `noexcept`.

**Why it matters:** Same `std::vector<Robot>` reallocation gate as M-006. Robot is small so the impact is minimal, but the discipline should match.

**Proposed fix:** add `noexcept` to move operations.

**Estimated effort:** S.

#### M-008 — `include/fields2cover/types/Geometries_impl.hpp:63,108`

**What:** `Iterator::Iterator(..., int nPos)` initializer list contains `m_poPrivate(new Private())`. If a future refactor adds another member after `m_poPrivate` that throws during its own init, the raw pointer leaks before `unique_ptr` takes ownership.

**Why it matters:** `std::make_unique<Private>()` is the exception-safe idiom. The current form is exception-unsafe in the presence of multi-member initializer lists.

**Proposed fix:**
```cpp
Iterator::Iterator(Geometries* poSelf, int nPos)
    : m_poPrivate(std::make_unique<Private>()) {
  m_poPrivate->m_poSelf = poSelf;
  m_poPrivate->m_nPos = nPos;
}
```
(same for `ConstIterator`).

**Estimated effort:** S.

#### M-009 — `include/fields2cover/types/Geometry_impl.hpp:51-62` (documentation gap)

**What:** `Geometry<T,R>::Geometry(const Geometry&)` is `= default`, which means the copy constructor performs a `shared_ptr` copy — the new `Geometry` **shares** the OGR data with the original. Mutation through one wrapper affects the other. `clone()` is the explicit deep-copy path. This model is applied to every derived wrapper (`Cell`, `Cells`, `LineString`, `LinearRing`, `Swath::path_`, etc.).

**Why it matters:** This is arguably the correct design (copy-on-write by caller convention), but it is not documented at the type level. A caller who writes `Swath a = b;` and then mutates `a.getPath()` will silently mutate `b` too. No memory corruption, but a semantic surprise that can produce subtly wrong route plans.

**Proposed fix:** add a doxygen block at the top of `Geometry.h` explaining the "copy = shallow share, clone = deep copy" convention, and list mutation methods that *do not* trigger a deep copy (`setPoint`, `addGeometry`, etc.).

**Estimated effort:** S (documentation only).

### Low

#### M-010 — `src/fields2cover/utils/transformation.cpp:246`

**What:** `new OGRSpatialReference()` is passed into the `std::unique_ptr` constructor directly.

**Why it matters:** No leak today (the 2-arg `unique_ptr(pointer, Deleter)` form with a function-pointer deleter is effectively nothrow in practice). It is, however, the only bare `new` in the library, and a `makeSpatialRef()` helper would make the allocation site self-documenting.

**Proposed fix:** optional — factor out a `makeSpatialRef()` helper; or leave as-is.

**Estimated effort:** negligible.

#### M-011 — `src/fields2cover/types/Cells.cpp:53-55`

**What:** `Cells::Cells(const Cell& c)` dereferences `this->data_->addGeometry(c.get())` without explicitly initializing `data_` in the body. It relies on the implicit base-class default constructor (`Geometry<>::Geometry()`) to populate `data_` first. Correct today, fragile under future refactors.

**Proposed fix:** use an explicit delegating constructor:
```cpp
Cells::Cells(const Cell& c) : Cells() {
  this->data_->addGeometry(c.get());
}
```

**Estimated effort:** trivial.

#### M-012 — `[[nodiscard]]` missing on pure queries / factories

**What:** `Cells::intersection`, `Cells::difference`, `Cells::unionOp`, `Cells::unionCascaded`, `Cells::convexHull`, `Cells::buffer`, `Path::discretizeSwath`, `Swath::clone`, `Field::clone`, `Route::clone`, `Swaths::clone`, `SwathsByCells::clone` all return a new value and have no side effect on `*this`. Calling any of them and ignoring the result is always a bug.

**Why it matters:** Low severity — obvious query-style APIs — but `[[nodiscard]]` catches the mistake at compile time for free.

**Proposed fix:** mechanical annotation sweep across the `types/` headers.

**Estimated effort:** S.

## Large-Type Ownership Matrix

| Type             | Rule-of-5 status                                     | Dtor              | Move noexcept? | Pass-by-value sites in public API                                | Notes                                                                |
|------------------|------------------------------------------------------|-------------------|----------------|------------------------------------------------------------------|----------------------------------------------------------------------|
| `Cells`          | implicitly via `Geometry<>::= default`               | `= default`       | computed yes   | none (always const&)                                             | Safe but fragile against base-class member changes.                  |
| `Swaths`         | **broken** — user dtor, no move ops declared         | user-declared     | **no**         | none                                                             | **M-002** — copies where it should move.                             |
| `SwathsByCells`  | **broken** — user dtor, no move ops declared         | user-declared     | **no**         | none                                                             | **M-003** — cascades through `Swaths`.                               |
| `Path`           | compiler-generated, all five; implicit               | implicit          | yes (vector)   | none                                                             | OK. Declaring `= default` explicitly would harden.                   |
| `Field`          | full rule-of-5 declared                              | user-declared     | **no**         | none                                                             | **M-006** — missing `noexcept` on moves.                             |
| `Route`          | compiler-generated; implicit                        | implicit          | inherited      | none                                                             | Cascades cost from `Swaths` (**M-002**).                             |
| `Swath`          | partial — copy ctor, virtual op=, no move ctor       | `virtual = default` | **no**       | `Swaths::emplace_back(const Swath&)` is const&                   | **M-005** — no move ctor; `virtual` op= is a latent slicing hazard.  |
| `Robot`          | full rule-of-5 declared                              | user-declared     | **no**         | none                                                             | **M-007** — missing `noexcept` on moves.                             |

No public API methods in `src/fields2cover/` or `include/fields2cover/` take
any of `F2CCells`, `F2CSwaths`, `F2CPath`, `F2CField`, `F2CRoute` **by value**
as a parameter. All such parameters are `const T&` or `T&` — the pass-by-value
anti-pattern does not exist in this codebase. That is a strong positive result.

Return-by-value is used throughout (`F2CPath PathPlanning::planPath(...)`,
`F2CRoute RoutePlannerBase::genRoute(...)`, etc.); NRVO handles these
correctly regardless of rule-of-5 status because a named return object is
elided. The rule-of-5 issues only bite on reallocation-driven moves inside
`std::vector<T>` and explicit `std::move(x)` sites.

## SWIG Boundary Notes

`swig/Fields2Cover.i` exposes the full `f2c::types::*` surface to Python
(and any other SWIG target). The boundary is **safe today** for three
reasons:

1. **No `%shared_ptr<T>` directive is in use.** Every wrapper type (`Cells`,
   `Swaths`, `Path`, `Field`, `Route`, ...) crosses SWIG by value. The Python
   object holds a C++ *copy* of the wrapper, which internally holds a
   `std::shared_ptr<OGRGeometry>`. OGR data is automatically shared across
   the wrappers; the wrapper itself is independently owned by the Python
   runtime.
2. **SWIG's default reference-return behavior is copy-out.** Methods like
   `Field::getField() -> Cells&`, `Route::getSwaths(size_t) -> Swaths&`,
   `Swaths::back() -> Swath&`, `Path::back() -> PathState&` return a C++
   reference, but SWIG's default typemap for non-primitive reference-returns
   allocates a new Python object *copied* from the referent. No Python
   object holds a dangling C++ address. (This can be overridden with
   `%naturalvar` or `%typemap(out) T&`, but neither is used anywhere in
   `Fields2Cover.i`.)
3. **`EmptyDestructor` is `%ignore`-d** (`Fields2Cover.i:35`), so the
   non-owning view constructor exposed in C++ (see **M-001**) is not
   reachable from Python. Python cannot construct a non-owning Cell wrapper.
   This is the only thing keeping M-001 from being a Python-visible footgun.

### Ignored symbols that matter for lifetime

| Symbol                                                                                   | Reason                                               |
|------------------------------------------------------------------------------------------|------------------------------------------------------|
| `f2c::types::Geometries::Iterator`, `ConstIterator`, `begin`, `end`                      | Iterator invalidation and `unique_ptr<Private>` wrap |
| `f2c::types::EmptyDestructor`                                                            | Non-ownership tag (see **M-001**)                    |
| `f2c::Transform::generateCoordTransf`, `createSptRef`, `createCoordTransf`               | Return `unique_ptr<OGR*, fn*>` which SWIG can't wrap |
| `f2c::rp::RoutePlannerBase::createShortestGraph`, `createCoverageGraph`, `computeBestRoute`, `transformSolutionToRoute` | Internal graph types                                 |
| `f2c::types::Swaths::operator[]`, `SwathsByCells::operator[]`, `Path::operator[]`        | `at()` used instead (returns copy in SWIG default)   |

### What could break the SWIG boundary

- Adding `%naturalvar f2c::types::Cells;` or similar — would make references
  return addresses instead of copies. Every reference-return in the public API
  would become a dangling-prone surface.
- Adding `%shared_ptr(f2c::types::X)` for any `X` that inherits from
  `Geometry<T,R>` — would require careful re-audit of the internal
  shared_ptr storage because SWIG's shared_ptr machinery is not aware of the
  inner `Geometry::data_` shared_ptr.
- Exposing the `(T*, EmptyDestructor)` / `(OGRGeometry*, EmptyDestructor)`
  constructors — would directly expose **M-001** to Python.

**Recommendation:** add a header comment to `swig/Fields2Cover.i` explicitly
stating these three invariants and pointing at this audit.

## Cross-references

- Findings in this report use `M-NNN` IDs (memory-audit).
- Static analysis findings (plan 06-01) use `F-NNN` IDs.
- Both feed into plan 06-04 (low-risk fixes — targets M-002, M-003, M-004,
  M-005 `noexcept`, M-006, M-007, M-008, M-011, M-012) and plan 06-05
  (high-risk triage — targets M-001, the `virtual operator=` portion of
  M-005, and the SWIG `%naturalvar` invariant).
- Scratch files: `build/audit/memory-survey.txt`, `build/audit/memory-review.md`
  (both gitignored; regenerable from this document's Method section).

## Appendix: files scanned

```
src/fields2cover/
├── decomposition/         — 0 new, 0 delete, 0 casts
├── headland_generator/    — 0 new, 0 delete, 0 casts
├── objectives/            — 0 new, 0 delete, 0 casts
├── path_planning/         — 0 new, 0 delete, 0 casts
├── route_planning/        — 0 new, 0 delete, 0 casts
├── swath_generator/       — 0 new, 0 delete, 0 casts
├── types/                 — 0 new, 0 delete, 0 casts (shared_ptr heavy)
└── utils/                 — 1 new (transformation.cpp:246), 0 delete

include/fields2cover/
├── decomposition/         — clean
├── headland_generator/    — clean
├── objectives/            — clean
├── path_planning/         — clean
├── route_planning/        — clean
├── swath_generator/       — clean
├── types/                 — 2 new (Geometries_impl.hpp PImpl), 1 delete (Geometry_impl.hpp:329 — **M-004**)
└── utils/                 — clean

swig/
├── Fields2Cover.i         — 341 lines, analyzed
├── optional.i             — typemap for std::optional — safe copy-out
└── python/                — Python-specific extension; reviewed
```
