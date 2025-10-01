`` `
#   FIRST element merge

FIRST: Float, Integer, Reference, String, Term.
Merges are last-write-wins, based on:

 1. ID seq, then
 2. ID src, then
 3. type, then
 4. value (type dependant).

 If 1-4 are equal, elements are identical.
```
a,
a,
a,
    ~,
a,

b,
a,
    ~,
b,

a b;
a:b:c;
    ~
a:b:c

a,
b,
    ~,
b,

1 @a-1,
2 @b-2,
3 @c-3,
~,
3@c-3,

1.23,
4.5,
~,
4.5,

1.23,
123,
~,
123,

"C" @c-3,
"B" @b-2,
"A" @a-1,
~,
"C"@c-3,

```
`` `

