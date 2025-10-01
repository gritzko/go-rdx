`` `
#   Eulerian containers (sets, maps) merge

The merge rules for an Eulerian set are rather straightforward.

Versions of the same set get merged, element by element.
```
{2:(5 6)},
{1:(2 33 4)},
{1:(2 33)},
~,
{1:(2 33 4),2:(5 6)},

{@alices-A 1 2 4}
{@alices-A 1 3 5}
    ~
{@alices-A 1 2 3 4 5}
```
That applies to no-id sets as well: they count as the same set.
```
{1},
{1},
    ~,
{1},

```
...entries merge...
```
{1 3},
{4 5},
    ~,
{1,3,4,5},

{4 5},
{1 3},
    ~,
{1,3,4,5},

```
...tuple entries merge...
```
{1:2},
{3:4},
~,
{1:2,3:4},

{1:2, 3:4},
{1:1, 3:5, 4:5},
~,
{1:2, 3:5, 4:5},

```
...nested sets merge...
```
{1:2, 3:{4,5}},
{1:2, 3:6, 7:8},
~,
{1:2, 3:{4,5}, 7:8},

```
The rules apply recursively. When we merge two versions of a set,
versions of its elements also get merged.
```
{1:2, 3:{4,5}},
{1:2, 3:{4:10}, 7:8},
~,
{1:2, 3:{4:10,5}, 7:8},

```
The above rules applied recursively again:
```
{},
{{}},
{{{}}},
~,
{{{}}},

{1:(2 3)},
{1:(2 3 44)},
{2:(5 6)},
~,
{1:(2,3,44),2:(5,6)},

{1:(2 3)},
{1:(2 3 4), 22:(5 6)},
~,
{1:(2 3 4),22:(5 6)},

{2:(5 6)},
{1:(2 33 4)},
{1:(2 33)},
~,
{1:(2 33 4),2:(5 6)},

{2:(5 6)},
{1:(2 3)},
{1:(2 3 444)},
{1:(2 3 444)},
{2:(5 6)},
{1:(2 3)},
~,
{1:(2 3 444),2:(5 6)},

```
Different sets (different ids) never get merged; it is either one or another.
```
{@alices-A0 1 2 3}
{@bobs-B0 4 5}
    ~
{@bobs-B0 4 5}

```
In the example above, the bigger id wins.
In case we want to replace one set by another, we change the revision
```
({@bobs-B 4 5})
(@20 {@alices-A 1 2 3})
    ~
(@20 {@alices-A 1 2 3})

```
Revision-envelopes with odd numbers count as tombstones.
So, set deletion looks like:  (TODO maybe skip the contents of a deleted E?)
```
{@alices-A 1 2 3}
{@alices-B}
    ~
{@alices-B 1 2 3}

```
Correspondingly, here is undeletion:
```
(@1 {@alices-A 1 2 3})
(@2 {@alices-A 1 2 3})
    ~
(@2 {@alices-A 1 2 3})

```
Again, enveloping is only relevant when our data model is not
a clean JSON-like tree, but we have to include arbitrary objects
by reference. If you want to keep things simple, use trees.
`` `

