# About
Trap is the core the xgo runtime. 

When a function is called, it's control flow will first be transferred to Trap.

Then Trap check if there is any interceptor, including system defined interceptors like `Mock` and `Trace`, as well as user defined interceptors by calling `Trap.AddInterceptor()`.

If any, it will then forward call to these interceptors, until all interceptors returned, or some interceptor returns `trap.ErrAbort` in the middle.

# `Inspect(f)`
the `trap.Inspect(fn)` implements a way to retrieve func info.
It has different internal paths for these function types:
- package level function
  - PC lookup
- struct method
  - `-fm` suffix check
  - `struct.method` proto existence check
  - dynamic retrieval
- interface method
  - `-fm` suffix check
  - interface proto existence check
  - dynamic retrieval
- closure
  - closure is registered using IR
  - PC lookup
- generic function
  - fullName parsing to de-generic
  - generic template existence check
  - no dynamic retrieval since no recv, but is legal to do so
- generic struct method
  - `-fm` suffix check
  - fullName parsing to de-generic
  - generic method template existence check
  - dynamic retrieval just for receiver
- generic interface method
  - to be supported