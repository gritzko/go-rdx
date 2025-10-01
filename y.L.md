`` `
#   This is a merge test set for Linear collections.

 1. Unstamped collections can only be appended to;
    items can be replaced.
    ```
    [1, 2, 3, 4]
    [1, 22@2, 3]
    ~
    [1, 22@2, 3, 4]
    ```
 2. Appends also work nicely for stamped collections.
    They are also more efficient; no need to cite the entire array.
    ```
    [a@10 b@30 c@50]
    [c@50 d@70]
    ~
    [a@10 b@30 c@50 d@70]
    ```
 3. Inserts are only possible into stamped collections.
    ```
    [a@10 b@30 c@50]
    [b@30 e@40 c@50 d@70]
    ~
    [a@10 b@30 e@40 c@50 d@70]
    ```
 4. With the numbering space depleted we have to use insert trains.
    ```
    [a@10 b@30 e@40 c@50 d@70]
    [a@10 aa@bob-20 ab@10 ac@210 ad@410 ae@610]
    ~
    [a@10 aa@bob-20 ab@10 ac@210 b@30 e@40 ad@410 c@50 ae@610 d@70]
    ```
 5. Although it may take some time for 64 bits to be depleted
    ```
    [0.0@20 1.0@30]
    [0.4@420]
    [0.2@220 0.3@320]
    [0.5@520 0.6@620]
    ~
    [0.0@20 0.2@220 1.0@30 0.3@320 0.4@420 0.5@520 0.6@620]
    ```
    ` ``
