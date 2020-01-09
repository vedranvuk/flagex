# Description

flagex is a command line parser with (optional) pacman-like style parsing.

It works with string values as its intermediate type. Reflect-based helpers are in the `reflag` package, which works, but has errorproofing left to do.


## Status

Functionality I had in mind is in. Additional helpers might be added. Not 100% on the API yet, but any further changes should be minor.

## Example

```

	// Define a sub.
	svpar := New()
	svpar.Req("addr", "a", "Listen address.", "HOSTNAME|IP", "0.0.0.0")
	svpar.Opt("tlsmode", "t", "specify TLS mode", "v1|v2|v3", "v3")

	// Define root.
	flags := New()
	flags.Sub("srvparams", "s", "Server parameters.", svpar)
	flags.Req("config", "c", "Specify config file.", "filename", "settings.json")
	flags.Switch("verbose", "v", "Verbose output.")

	if err := flags.Parse(os.Args[1:]); err != nil {
		// Use flags:
		// flags.Parsed()
		// flags.ParseMap()
		// flags.Print()
		// ...
	}

	// Example command line:
	//
	// Same effect:
	// --verbose --config mysettings.json --srvparams --tlsmode v3 -a 127.0.0.1
	// -v -c mysettings.json -st v3 --addr 127.0.0.1

	// flags.Print() yields:
	//
	//	-s   --srvparams           Server parameters.     
	//		-a                    --addr <HOSTNAME|IP>   Listen address.    
	//		-t                    --tlsmode <v1|v2|v3>   specify TLS mode   
	//	-c   --config <filename>   Specify config file.   
	//	-v   --verbose             Verbose output.    	

```

## License

MIT.