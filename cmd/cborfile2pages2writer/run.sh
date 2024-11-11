#!/bin/sh

export ENV_PAGED_CBOR_FILENAME=./sample.d/input.paged.cbor
export ENV_CBOR_PAGE_SIZE=Tiny

./cborfile2pages2writer |
	python3 \
	-m uv \
	tool \
	run \
	cbor2 \
	--sequence
