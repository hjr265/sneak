# Sneak

Sneak is a program that can hide a file in other files.

Container formats supported:

- ZIP

_Yes, just one for now._

## Usage

Install Sneak by running the following command:

``` sh
go install github.com/hjr265/sneak@latest
```

### Example

``` sh
# Make a ZIP file
echo 'Hello World' > nothamlet.txt
zip archive.zip nothamlet.txt

# Make a secret file
echo 'KeyboardCat' > secret.txt

# Add secret.txt to archive.zip
sneak archive.zip secret.txt

# Remove the  secret file
rm secret.txt
cat secret.txt # Won't work. We have just removed it.

# Extract hidden file from archive.zip
sneak -x archive.sneak.zip
cat secret.txt # Will print 'KeyboardCat'.
```
