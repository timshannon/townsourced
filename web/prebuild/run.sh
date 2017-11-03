#!/bin/bash
gobble build out

mv out/emoji.json ../src/js/ts/emojiData.js

rm -rf out
