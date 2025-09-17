#!/bin/sh
# simple editor script for testing
# replaces foo with bar and baz with qux
mv "$1" "$1.bak"
sed -e 's/foo/bar/g' -e 's/baz/qux/g' "$1.bak" > "$1"
rm "$1.bak"
