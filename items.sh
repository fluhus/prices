if [ $# != 1 ]; then
	echo Usage: 1>&2
	echo 'items.sh <file>' 1>&2
	exit 1
fi

file=tmp.xml

if [ ${1: -3} == .gz ]; then
	gunzip -c $1 > $file
elif [ ${1: -4} == .zip ]; then
	unzip -p $1 > $file
else
	cp $1 $file
fi

encoding=`grep -P -o 'encoding=".*?"' $file`
encoding=${encoding:10: -1}

if [ $encoding. != . ] && [ $encoding != utf-8 ]; then
	iconv -f $encoding -t utf-8 < $file > $file.2
	if [ $? != 0 ]; then
		rm $file $file.2
		exit 2
	fi
	mv $file.2 $file
fi

sed -r 's/encoding=".*?"/encoding="utf-8"/' < $file > $file.2
mv $file.2 $file

/cs/grad/amitlavon/Desktop/prices/bin/items < $file
code=$?
rm -f $file $file.2
exit $code
