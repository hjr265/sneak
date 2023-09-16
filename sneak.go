package main

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hjr265/sneak/archive/zip"
)

var (
	flagExtract = flag.Bool("x", false, "Extract mode")
	flagOutput  = flag.String("o", "", "Output file")
)

func main() {
	flag.Parse()

	if *flagExtract {
		f, err := os.Open(flag.Arg(0))
		catch(err)
		fi, err := f.Stat()
		catch(err)

		sr, err := zip.NewReader(f, fi.Size())
		catch(err)

		outname := *flagOutput
		if outname == "" {
			outname = sr.Header().Name
		}

		o, err := os.Create(outname)
		catch(err)
		_, err = io.Copy(o, sr)
		catch(err)
		err = o.Close()
		catch(err)

		return
	}

	f, err := os.Open(flag.Arg(0))
	catch(err)
	fi, err := f.Stat()
	catch(err)

	outname := *flagOutput
	if outname == "" {
		ext := filepath.Ext(flag.Arg(0))
		outname = strings.TrimSuffix(flag.Arg(0), ext) + ".sneak" + ext
	}

	o, err := os.Create(outname)
	catch(err)
	m, err := os.Open(flag.Arg(1))
	catch(err)
	sw, err := zip.NewWriter(o, f, fi.Size())
	catch(err)
	err = sw.SneakFile(m)
	catch(err)
	err = o.Close()
	catch(err)
}

func catch(err error) {
	if err != nil {
		panic(err)
	}
}
