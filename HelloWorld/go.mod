module example.com/hello

go 1.16

require (
	example.com/mymodule v0.0.0-00010101000000-000000000000
	rsc.io/quote v1.5.2
)

replace example.com/mymodule => ../MyModule