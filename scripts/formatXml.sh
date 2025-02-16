retval=0
tmpfile=$(mktemp)
echo "Using temp file $tmpfile"
while read -r file; do
    echo "Proceccing file $file"
    xmllint --format $file --output $tmpfile
    diff $file $tmpfile > /dev/null
    if [ $? -ne 0 ]; then
        echo "$file was reformatted"
        retval=1
    fi
    mv $tmpfile $file
done < <(find "." -type f \( -name "*.xml" -o -name "*.xsd" -o -name "*.musicxml"  \))
echo "Exiting with status $retval"
exit $retval
