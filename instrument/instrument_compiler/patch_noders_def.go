package instrument_compiler

const NoderFiles_1_17 = `
// auto gen
if /*os.Getenv("XGO_COMPILER_SYNTAX_REWRITE_ENABLE")=="true"*/ true {
	files := make([]*syntax.File, 0, len(noders))
	for _, n := range noders {
		files = append(files, n.file)
	}
	xgo_syntax.AfterFilesParsed(files, func(name string, r xgo_io.Reader) *syntax.File {
		p := &noder{
			err: make(chan syntax.Error),
		}
		fbase := syntax.NewFileBase(name)
		file, err := syntax.Parse(fbase, r, nil, p.pragma, syntax.CheckBranches)
		if err != nil {
			e := err.(syntax.Error)
			p.error(e)
			return nil
		}
		p.file = file
		noders = append(noders, p)

		// move to head
		n := len(noders)
		for i:=n-1;i>0;i--{
			noders[i]=noders[i-1]
		}
		noders[0]=p

		return file
	})
}
`

const NoderFiles_1_20 = `
// auto gen
if os.Getenv("XGO_COMPILER_SYNTAX_REWRITE_ENABLE")=="true" {
	files := make([]*syntax.File, 0, len(noders))
	for _, n := range noders {
		files = append(files, n.file)
	}
	xgo_syntax.AfterFilesParsed(files, func(name string, r xgo_io.Reader) *syntax.File {
		p := &noder{}
		fbase := syntax.NewFileBase(name)
		file, err := syntax.Parse(fbase, r, nil, p.pragma, syntax.CheckBranches)
		if err != nil {
			e := err.(syntax.Error)
			base.ErrorfAt(p.makeXPos(e.Pos), "%s", e.Msg)
			return nil
		}
		p.file = file
		noders = append(noders, p)

		// move to head
		n := len(noders)
		for i:=n-1;i>0;i--{
			noders[i]=noders[i-1]
		}
		noders[0]=p

		return file
	})
}
`

const NoderFiles_1_21 = `
// auto gen
if os.Getenv("XGO_COMPILER_SYNTAX_REWRITE_ENABLE")=="true" {
	files := make([]*syntax.File, 0, len(noders))
	for _, n := range noders {
		files = append(files, n.file)
	}
	xgo_syntax.AfterFilesParsed(files, func(name string, r xgo_io.Reader) *syntax.File {
		p := &noder{}
		fbase := syntax.NewFileBase(name)
		file, err := syntax.Parse(fbase, r, nil, p.pragma, syntax.CheckBranches)
		if err != nil {
			e := err.(syntax.Error)
			base.ErrorfAt(m.makeXPos(e.Pos), 0,"%s", e.Msg)
			return nil
		}
		p.file = file
		noders = append(noders, p)

		// move to head
		n := len(noders)
		for i:=n-1;i>0;i--{
			noders[i]=noders[i-1]
		}
		noders[0]=p

		return file
	})
}
`
