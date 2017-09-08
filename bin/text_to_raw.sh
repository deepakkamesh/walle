#!/bin/bash

flite -voice "slt" -o /tmp/out.wav -t "$1"
sox /tmp/out.wav --bits 16 --encoding signed-integer --endian little output.raw
rm /tmp/out.wav
