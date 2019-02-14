#!/bin/bash
GHOSTSCRIPT_BINARY="gs"
FONTPATH="."
OCR_LANG="eng"

INPATH="$1"
#echo "INPATH: ${INPATH}"
INFILE=$(basename "${INPATH}")
#echo "INFILE: $INFILE"
FILENAME="${INFILE%.*}"
#echo "FILENAME: ${FILENAME}"
DSTPATH="./DST/$FILENAME"
#echo $DSTPATH
WORKINGDIR="${DSTPATH}/TMP"
#echo $WORKINGDIR
FINAL_OUTFILE="$DSTPATH/$FILENAME.pdf"
#echo $FINAL_OUTFILE
OUTFILE="$WORKINGDIR/$FILENAME.pdf"
#echo $OUTFILE
OUTFILE_REPAIRED="${WORKINGDIR}/${FILENAME}_REPAIRED.pdf"
#echo $OUTFILE_REPAIRED
OUTFILE_LINEARIZED="${WORKINGDIR}/${FILENAME}_LINEARIZED.pdf"
#echo $OUTFILE_LINEARIZED

mkdir -p "$WORKINGDIR"
printf "=================================================\n"
printf "Preparing: ${FILENAME}"
printf "=================================================\n"

### Show file info
printf "\nFile Info:\n"
printf "=================================================\n"
exiftool -all:all "$INFILE"

### Strips editing restrictions (password) by running through ghostscript
### FIXME: This sometimes does some unwelcome font replacement
printf "\nStripping Edit Restrictions:\n"
printf "=================================================\n"
$GHOSTSCRIPT_BINARY -dSAFER -dBATCH -dNOPAUSE -sDEVICE=pdfwrite -sPDFPassword= -dPDFSETTINGS=/prepress -sFONTPATH="$FONTPATH" -dPassThroughJPEGImages=true -sOutputFile="$OUTFILE" "$INPATH"

## Repair anything broken
printf "\nRepairing...\n"
printf "=================================================\n"
pdftk "$OUTFILE" output "$OUTFILE_REPAIRED"

## Strip metadata from the PDF
printf "\nStripping metadata...\n"
printf "=================================================\n"
exiftool -all= "$OUTFILE_REPAIRED"
exiftool -all:all "$OUTFILE_REPAIRED"

# Linearize (optimise them for fast web loading and removes any orphan data.)
printf "\nOptimizing...\n"
printf "=================================================\n"
qpdf --linearize "$OUTFILE_REPAIRED" "$OUTFILE_LINEARIZED"

# Remove embedded metadata
printf "\nRemove embedded metadata...\n"
printf "=================================================\n"
exiftool -extractEmbedded -all:all "$OUTFILE_LINEARIZED"

# Compressing
printf "\nSaving clean copy (with compression)...\n"
printf "=================================================\n"
pdftk "$OUTFILE_LINEARIZED" output "$FINAL_OUTFILE" compress

# The next two lines are a "compression" scheme someone on the internet found, it is extremely slow
# but in some cases gives large gains. Not reliable enough to leave in though.
#pdf2ps "$FINAL_OUTFILE" "${WORKINGDIR}/${FILENAME}.ps"
#ps2pdf "${WORKINGDIR}/${FILENAME}.ps" "$FINAL_OUTFILE"

# Extracting existing text
printf "\nExtracting existing text...\n"
printf "=================================================\n"
pdftotext "$OUTFILE_LINEARIZED" "$DSTPATH/$FILENAME-extracted_text.txt"

# Extract image
printf "\nExtracting first page as image... (This might take a while)\n"
printf "=================================================\n"
convert -density 708x708 "$OUTFILE_LINEARIZED[0]" "$DSTPATH/$FILENAME-cover.jpg"

# OCR
printf "\nOCR...\n"
printf "=================================================\n"
printf "\n\n******* \n ...Step 1: Generating PNGs (This might take a while)\n"
cd "$WORKINGDIR"
INFILE="../TMP/${FILENAME}_LINEARIZED.pdf"
pdftoppm -png "$INFILE" PDF
printf "\n\n******* \n ...Step 2: OCRing\n"
a=1; for i in $(ls -v *.png) ; do echo "$i page_${a}.txt" ; tesseract -l $OCR_LANG $i page_${a}.txt ; let a=a+1 ; done
for i in $(ls -v page_*.txt) ; do cat $i ; done > "../${FILENAME}-ocr_text.txt"
cd ..

# Cleanup
rm -rf ./TMP
