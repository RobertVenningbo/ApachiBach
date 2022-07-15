
## Welcome to Apachi – an Accountable, Secure and Fair Approach to Conference Paper Submission, an implementation.

This project is the bachelor thesis of Robert Venningbo & Philip Bilbo regarding an implementation of the master thesis *"Apachi – an Accountable, Secure and Fair Approach to Conference Paper Submission"* by Yoav Schwartz & Nicolai Strøm Steffensen.

  

### How to run the protocol.

  

1. Clone the repository

2. Open a terminal and navigate to the root folder of the project

- Type ```go run .\Quickstart\main.go```  *(ONLY TESTED ON WINDOWS)*
	- *Sample output*

```
PS C:\Kode\ApachiBach> go run .\Quickstart\main.go
Welcome to Apachi, please input the amount of agents.
Please input the amount of submitters: 
2
Please input the amount of reviewers: 
2

PC:             http://localhost:4000
Submitter 3000: http://localhost:3000
Submitter 3001: http://localhost:3001
Reviewer 5000:  http://localhost:5000
Reviewer 5001:  http://localhost:5001 
```


### **If that doesn' work open a series of terminals and write:**
*(links will be the same format as the above example)*

-  ```go run .\main.go *agent* *port* ```
  

***For example:***

-  ```go run .\main.go pc 4000 ```
-  ```go run .\main.go submitter 3000 ```
-  ```go run .\main.go submitter 3001 ```
-  ```go run .\main.go reviewer 5000```
-  ```go run .\main.go reviewer 5001 ```