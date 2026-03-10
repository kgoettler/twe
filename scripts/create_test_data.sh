#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
DBDIR="${SCRIPT_DIR}/../pkg/timewarrior/testdata/db"

# Main
if [ ! -d "$DBDIR" ]; then
    mkdir -p $DBDIR
fi
export TIMEWARRIORDB=$(readlink -f "$DBDIR")

# Make sure TIMEWARRIORDB is set before we start deleting stuff
if [ "$TIMEWARRIORDB" == "" ]; then
    echo "error: TIMEWARRIORDB is not being set properly in this script!"
    exit 1;
fi

# Remove existing data, extensions, etc.
rm -rf $TIMEWARRIORDB/data $TIMEWARRIORDB/extensions

# Initialize
echo "bootstrapping timewarriordb..."
timew > /dev/null
echo "...done"

echo "bootstrapping test data..."
# Create test data
# 2025-12 for VHS demos (because it's a nice and round and starts on a Monday)
# 2026-01 for all other tests (because it's nice and round for writing)
MONTHS=("202601" "202512")
for month in "${MONTHS[@]}"; do
    for day in {1..7}; do
        day=$(printf '%02d' $day)
        timew track "${month}${day}T000000" - "${month}${day}T060000" "Sleep" > /dev/null
        timew track "${month}${day}T060000" - "${month}${day}T070000" "Shower" > /dev/null
        timew track "${month}${day}T070000" - "${month}${day}T080000" "Breakfast" > /dev/null
        timew track "${month}${day}T080000" - "${month}${day}T090000" "Commuting to Work" > /dev/null
        timew track "${month}${day}T090000" - "${month}${day}T170000" "Work" > /dev/null
    done
done
echo "...done"

# Add data for today
today=$(date +%Y%m%d)
timew track "${today}T000000" - "${today}T060000" "Sleep" > /dev/null
timew track "${today}T060000" - "${today}T070000" "Shower" > /dev/null
timew track "${today}T070000" - "${today}T080000" "Breakfast" > /dev/null
timew track "${today}T080000" - "${today}T090000" "Commuting to Work" > /dev/null
timew track "${today}T090000" - "${today}T093000" "Admin" > /dev/null
timew track "${today}T093000" - "${today}T100000" "Meeting" > /dev/null
timew track "${today}T100000" - "${today}T103000" "JIRA-100" > /dev/null
timew track "${today}T103000" - "${today}T110000" "JIRA-100" "Meeting" > /dev/null
timew track "${today}T110000" - "${today}T120000" "JIRA-101" > /dev/null
timew track "${today}T120000" - "${today}T170000" "JIRA-102" > /dev/null

# Create "echo" extension
echo "bootstrapping extensions..."
outfile="${TIMEWARRIORDB}/extensions/echo.sh"
touch "$outfile"
outfile=$(readlink -f "$outfile")
cat <<EOF > "$outfile"
#!/bin/bash
cat -
EOF
chmod +x "$outfile"
echo "...done"