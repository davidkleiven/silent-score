TMP_DIR=$(mktemp -d)
DEST=$TMP_DIR/musicxml.zip

echo "Downloading latest version"
wget -O $DEST https://github.com/w3c/musicxml/releases/download/v4.0/musicxml-4.0.zip

UNZIPPED_DIR=$TMP_DIR/musicxml
unzip $DEST -d $UNZIPPED_DIR

echo "Generating Go data models"

echo "Creating data models for $name"
xgen -i $UNZIPPED_DIR/schema/musicxml.xsd -o internal/musicxml/musicxml.go -l Go -p musicxml
xgen -i $UNZIPPED_DIR/schema/xlink.xsd -o internal/musicxml/xlink.go -l Go -p musicxml
xgen -i $UNZIPPED_DIR/schema/xml.xsd -o internal/musicxml/xml.go -l Go -p musicxml

echo "Deleting temporary directory"
rm -r $TMP_DIR
