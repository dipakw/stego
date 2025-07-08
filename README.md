# Stego

Stego is a steganography library written in Go. Steganography allows you to hide data within other files, like images or documents, so they appear unchanged to the naked eye. This library allows you to embed secret messages and extract hidden data from files without anyone knowing the data is there.

# Supported file types

1. PNG
2. BMP
3. TIFF
4. WAV
5. WebP

# Installation

```bash
go get github.com/dipakw/stego
```

# Example

## Embed secret data into a file

```go
package main

import (
	"github.com/dipakw/stego"
	"fmt"
)

func main() {
	seed := []byte{0x00, 0x0e, 0x03, 0x99}

	file, err := stego.New("path/to/image.png", &stego.Opts{
		RandSeed:  seed,    // Used to randomize the data across the file.
		UseSpace:  0.5,     // How much of the capacity do you want to use? Max is 1.
	})

	// Get the storage capacity in bytes of the file.
	// The .Cap() method is affected by the "UseSpace" option.
	fmt.Println("Storage capacity is", file.Cap(), "bytes.")

	// Prepare the data to be written.
	data := []byte("my-secret-data")

	// Write the data.
	n, err := file.Write(data)

	if err != nil {
		fmt.Println("Failed to write:", err)
		return
	}

	fmt.Println("Written", n, "bytes.")

	// Save the updated file.
	err = file.Save("path/to/new.png")

	if err != nil {
		fmt.Println("Failed to save the file:", err)
		return
	}

	fmt.Println("Saved successfully.")
}
```

## Read embedded data from a file

```go
package main

import (
	"github.com/dipakw/stego"
	"fmt"
)

func main() {
	seed := []byte{0x00, 0x0e, 0x03, 0x99}

	file, err := stego.New("path/to/new.png", &stego.Opts{
		RandSeed:  seed,    // Used to randomize the data across the file.
		UseSpace:  0.5,     // How much of the capacity do you want to use? Max is 1.
	})

	data := make([]byte, 14)

	n, err := file.Read(data)

	if err != nil {
		fmt.Println("Failed to read:", err)
		return
	}

	fmt.Println("Data:", string(data[:n]))
}
```

# TODO
- Implement encryption (currently the Encrypted and SecretKey options are defined but not used).

# Disclaimer

This software is provided for good purposes only. Users are liable for any harmful or illegal use of this library. The authors and contributors are not responsible for any misuse of this software. Use at your own risk and ensure compliance with all applicable laws.