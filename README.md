## Welcome to Apachi – an Accountable, Secure and Fair Approach to Conference Paper Submission, an implementation.
This project is the  bachelor thesis of Robert Venningbo & Philip Bilbo regarding an implementation of the master thesis *"Apachi – an Accountable, Secure and Fair Approach to Conference Paper Submission"* by Yoav Schwartz & Nicolai Strøm Steffensen.

### How to run the protocol.

 1. Clone the repository
 2. Open a terminal and navigate to the root folder of the project
 - type ```go run .\Quickstart\main.go ```  *(ONLY TESTED ON WINDOWS)*
 
 If that doesn' work open a series of terminals and write:
 - ```go run .\main.go *agent* *port* ```

***For example:***
 - ```go run .\main.go pc 4000 ```
 - ```go run .\main.go submitter 3000 ```
 - ```go run .\main.go submitter 3001 ```
 - ```go run .\main.go reviewer 5000```
 - ```go run .\main.go reviewer 5001 ```
 
