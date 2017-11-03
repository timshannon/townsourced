echo "Extract release file"
if ! 
then
	echo failed to extract release file
	exit 1
fi

echo "Setting permissions"
if ! chown -R root:root release
then
	echo failed to set root ownership
	exit 1
fi

echo "stopping townsourced service"
if ! systemctl stop townsourced.service
then
	echo failed to stop townsourced service
	exit 1
fi

echo "preserving large bin files"
if ! mv /usr/local/share/townsourced/web/bin release/usr/local/share/townsourced/web/bin
then
	echo failed to preserve large files
	exit 1
fi

echo "updating static web data"
if ! rm -rf /usr/local/share/townsourced/
then
	exit 1
fi
if ! mv release/usr/local/share/townsourced /usr/local/share/townsourced
then
	exit 1
fi

echo "updating townsourced executable"
if ! rm -rf /usr/local/bin/townsourced
then
	exit 1
fi
if ! mv release/usr/local/bin/townsourced /usr/local/bin/townsourced
then
	exit 1
fi

echo "starting townsourced service"
if ! systemctl start townsourced.service
then
	echo failed to start townsourced service
	exit 1
fi

echo "Cleanup old release folder"
rm -rf release
rm release.tar.gz

