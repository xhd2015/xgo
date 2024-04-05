# Trace
Trace allows visualize generated trace with a crafted UI.

# Usage
```sh
xgo tool trace path/to/Trace.json
```

for example:
```sh
# assuming starting at project root
cd runtime
# run test, generate TestUpdateUserInfo.json
xgo test ./stack_trace

cd ..
xgo tool trace ./runtime/test/stack_trace/TestUpdateUserInfo.json
```