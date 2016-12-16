# Golang implementation of simg2img

## Overview

A go (golang) library for decoding android sparse image formatted files into raw images.

## Examples

*For a full example see example utility in ./cmd/*

Using the writer. The performance of the writer is much better than the performance of the reader, as we can seek-past any "don't care" chunks in the sparse format, instead of writing random data.

```go
func simg2img(in_filename, out_filename string) error {
    in, err := os.Open(in_filename)
    if err != nil {
        return err
    }
    defer in.Close()

    out, err := os.Create(out_filename)
    if err != nil {
        return err
    }
    defer out.Close()

    writer := sparse.Simg2imgWriter(out)
    _, err = io.Copy(writer, in)
    return err
}
```

Using the reader (slower!! Read Above!)

```go
func simg2img_alt(in_filename, out_filename string) error {
    in, err := os.Open(in_filename)
    if err != nil {
        return err
    }
    defer in.Close()

    out, err := os.Create(out_filename)
    if err != nil {
        return err
    }
    defer out.Close()

    reader, err := sparse.Simg2imgReader(in)
    if err != nil {
        return err
    }

    _, err = io.Copy(out, reader)
    return err
}
```

## License
See LICENSE file.
