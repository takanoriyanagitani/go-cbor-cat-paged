#!/bin/sh

export ENV_PAGED_CBOR_FILENAME=./sample.d/pages.small.cbor
export ENV_CBOR_PAGE_SIZE=Small

indices=./sample.d/indices.dat

cat "${indices}" |
./indexedcbors2writer |
	python3 \
	-m uv \
	tool \
	run \
	cbor2 \
	--sequence
