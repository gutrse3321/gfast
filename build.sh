DIST="./dist"

if [ ! -d "$DIST"]; then
  mkdir "$DIST"
  mkdir $DIST/swagger
else
  rm -rf $DIST
  mkdir "$DIST"
  mkdir "$DIST"/swagger
fi

echo "go build now ..."
go build -o server

echo "copy to \"dist\" folder"
cp ./server $DIST
rm ./server
cp ./swagger/swagger.json $DIST/swagger/

clear
echo "all over down"

