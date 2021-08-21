# dzhigit
Git implementation using Golang 

## Motivation
I am using git in a daily basics but know nothing about the internal data structures used by git. It took Linus a few weeks to implement git , how many months it will take me ? Will see...

## TODO
1. [X] cat-file
2. [o] checkout
    1. [X] Change branch
    2. [X] Change files content
    3. [ ] Add security check
3. [X] commit-tree
4. [X] hash-object 
5. [X] init 
6. [ ] log - just print all commits with messages starting from head
7. [X] ls-tree 
8. [ ] merge - To consider
9. [ ] rm 
10. [ ] tag
11. [X] index
12. [X] write-tree
13. [X] update-ref
    
## How it works
### Blobs
The blobs are basically implemented in the same way they work in original git.
Blob file contains two peaces of information
1. Header
2. Zipped data

Here is the format for blob file `type length\x00zipped_data` where
1. **type** - the object type followed by space (blob,tree etc)
2. **length** - the length of unzipped data(without header)
3. **\x00** - null character
4. **zipped_data** - zipped representation of original data(without header)

### Index file
In order to implement staging area git uses [Index file](https://mincong.io/2018/04/28/git-index/) but original Index format is a complex binary file and it contains too much information. The Index file used by `dzhigit` is a simple text file with a list of entries in the following format
```
Mode C_time M_time sha1-hash F_name
```
Where
1. **Mode** - is a file mode(100644 - normal file,100755 - executable)
2. **C_time** - file's creation time in unix
3. **M_time** - file's last modification time in unix
4. **sha1-hash** - file's hash generated by `dzhigit hash-object command`
5. **F_name** - file's name

## Working On


## Resources
1. [Git magic](http://www-cs-students.stanford.edu/~blynn/gitmagic/ch01.html)
2. [Write yourself a Git!](https://wyag.thb.lt/)
