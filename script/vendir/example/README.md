# Verify vendir create
```sh
# cd to current directory if not yet
cd ./script/vendir/example

# update dependencies
# you may need go1.22
(cd src && go mod vendor)

# remove target vendor
rm -rf ./internal/third_party_vendir

# update vendor
go run github.com/xhd2015/xgo/script/vendir create ./src ./internal/third_party_vendir

# check result
go run ./test
# output:
#    github.com/xhd2015/xgo/script/vendir/example/test
```