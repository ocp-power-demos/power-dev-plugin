#!/usr/bin/env bash

# This script builds the container-device-interface spec per https://github.com/cncf-tags/container-device-interface/blob/main/SPEC.md

#echo "Discovering Target Devices to load into cdi.json"
TARGET_DEVICES=()
for BLOCK_DEVICE in $(lsblk --json -s -f --paths --include 253 | jq -r '.blockdevices[] | select(.fstype != "xfs" and select(.label == null or .label != "boot") and select(.name | startswith("/dev/nvme") | not)) | select(.name | startswith("/dev/mapper/")) | .name')
do
    BLOCK_DETAIL=$(lsblk "${BLOCK_DEVICE}" --include 253 --noheadings --json -s -f --paths)
    if [ ! -z "${BLOCK_DETAIL}" ]
    then
        # Block, Children (mapper,device)
        for COMPONENT in $(echo "${BLOCK_DETAIL}" | jq -r '[.blockdevices[]?, .blockdevices[]?.children[]?,.blockdevices[]?.children[]?.children[]?] | .[]?.name')
        do
            TARGET_DEVICES+=("${COMPONENT}")
            #echo ${COMPONENT}
        done
    fi
done

# Header needs v0.6.0
echo '{
    "cdiVersion": "0.6.0",
    "kind": "ibm.com/devices",
    "annotations": {
        "vendor" : "IBM",
        "dbms-devices": "container-visibility"
    },
    "devices": [
        {
            "name": "all",
            "containerEdits": {
				"env": ["injected_devices=true"],
				"deviceNodes": ['

first=true

for DEVICE in "${TARGET_DEVICES[@]}"
do
    # Print a comma before each entry except the first
    if [ "$first" = true ]; then
      first=false
    else
      echo ','
    fi

  # symbolic link... and it exists
  SAVED_DEVICE=${DEVICE}
  if [ -h "${DEVICE}" ]
  then
    TMP_DEVICE=$(readlink -f ${DEVICE})
    DEVICE=${TMP_DEVICE}
  fi

  # Grab the raw "ls -l" line
  line=$(ls -l "$DEVICE")

  # Example ls -l output for a device might be:
  # crw-rw---- 1 root tty 4, 1 Jan 16 10:00 /dev/tty1
  # Fields:       1    2   3   4   5,6 ...
  #
  # We want the 5th field, which contains "<major>,<minor>"
  major=$(echo "$line" | awk '{print $5}' | cut -d, -f1)

  minor=$(lsblk $DEVICE --noheadings --paths | sed 's|:| |g' | grep $DEVICE | awk '{print $3}')

  if [ -z "${minor}" ]
  then
    minor=$(echo $DEVICE | sed 's|-| |g' | awk '{print $NF}' | tail -n 1)
  fi 

  echo " \
                    { \
                        \"path\": \"$DEVICE\", \
                        \"major\": $major, \
                        \"minor\": $minor, \
                        \"permissions\": \"rwm\" \
                    } \
    "
done

# Close the JSON
echo '
                ]
            }
        }
    ]
}'