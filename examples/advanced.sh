#!/bin/bash

# Advanced Bash script with functions, pipes, and subshells
# This demonstrates more complex features that bash2go should handle

# Function that uses pipes internally
process_data() {
    local input="$1"
    echo "$input" | grep -v "^#" | sort | uniq
}

# Function that uses a subshell
in_temp_dir() {
    local output
    output=$(
        cd /tmp
        pwd
        ls -la | grep "^d" | wc -l
    )
    echo "Found $output directories in /tmp"
}

# Function that uses conditionals and loops
process_files() {
    local dir="$1"
    local count=0
    
    if [ ! -d "$dir" ]; then
        echo "Error: $dir is not a directory"
        return 1
    fi
    
    for file in "$dir"/*; do
        if [ -f "$file" ]; then
            echo "Processing $file"
            count=$((count + 1))
        fi
    done
    
    echo "Processed $count files"
    return 0
}

# Main script execution

# Using pipes
echo "Starting advanced script"
ls -la | grep "^d" | awk '{print $9}' | while read -r dirname; do
    echo "Directory: $dirname"
done

# Using a here document with a pipe
cat << EOF | grep "important"
This is a test
This line has important information
Another test line
EOF

# Using process substitution
diff <(ls -la) <(ls -la /tmp)

# Function calls
echo "Sample data:" > /tmp/sample.txt
echo "# Comment line" >> /tmp/sample.txt
echo "Data line 1" >> /tmp/sample.txt
echo "Data line 2" >> /tmp/sample.txt
echo "Data line 1" >> /tmp/sample.txt

process_data "$(cat /tmp/sample.txt)"
in_temp_dir
process_files "/tmp"

# Using command substitution in a complex way
echo "Current date: $(date "+%Y-%m-%d")"
echo "Files in current directory: $(ls | wc -l)"

# Using multiple pipes and redirections
find . -type f -name "*.sh" | xargs grep "echo" | tee /tmp/output.txt | wc -l

# Cleanup
rm -f /tmp/sample.txt /tmp/output.txt

exit 0 