retval=0
tmpfile=$(mktemp)
tmpfile2=$(mktemp)
echo "Using temp file $tmpfile"
while read -r file; do
    echo "Proceccing file $file"
    go run scripts/sortMusicxml/sort_musicxml.go $file $tmpfile
    xmllint --dropdtd $tmpfile --output $tmpfile
    xmllint --format $tmpfile --output $tmpfile
    xmllint --exc-c14n $tmpfile > $tmpfile2
    sed -i -e '$a\' $tmpfile2
    diff $file $tmpfile2 > /dev/null
    if [ $? -ne 0 ]; then
        echo "$file was reformatted"
        retval=1
    fi
    mv $tmpfile2 $file
done < <(find "." -type f \( -name "*.xml" -o -name "*.xsd" -o -name "*.musicxml"  \))
rm $tmpfile
echo "Exiting with status $retval"
exit $retval
