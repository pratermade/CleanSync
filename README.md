# S3sync

s3sync

## Description

A little personal project I use to keep my videos backed up to S3. I've spent amy hours processing my video library so I thought is was worth a bit of effort and a few buck a month to upload them to the S3 Glacier Deep Archive in the case that anything happens to my local media.

## Getting Started

Download and build with Go 1.21. Uses sqlite so you will need the gcc libries to build. Download GCC from (http://tdm-gcc.tdragon.net/download)[http://tdm-gcc.tdragon.net/download]

### Dependencies

Uses sqlite so you will need the gcc libries to build. Download GCC from (http://tdm-gcc.tdragon.net/download)[http://tdm-gcc.tdragon.net/download]
Was built targeting windows 10, but I don't see any reason it would not work on Linux or Mac

### Installing

After it is built, copy it to where you want it to live. It will create a manifest.db in that directory when ran. This is where it catalogs the files that it backs up.

### Executing program


```
s3sync -help # for information
.\s3sync.exe sync -path=x:\videos -bucket=my-backup-bucket -filter=mkv -filter=mp4 -deep # example command
```

```
NAME:
   s3sync - Sync with the provided s3 bucket

USAGE:
   s3sync [global options] command [command options]

COMMANDS:
   sync     upload new files to the provided bucket
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help
```

*Subcommands* 

```
NAME:
   s3sync sync - upload new files to the provided bucket

USAGE:
   s3sync sync [command options]

OPTIONS:
   --path value, -p value                                 The source (local) folder to sync with S3
   --bucket value, -b value                               The name of the bucket to sysnc to
   --filter value, -f value [ --filter value, -f value ]  file types to filter for. Can be specified multiple times for multiple file types.
   --deep, -d                                             deep archive in S3 (default: false)
   --help, -h                                             show help
```

## Version History

* 0.1
    * Initial Release
* 0.2
    * Added file splitting for files greater than 4GB
* 0.3
   * Cleaned up cli output. Now it is simpler, but accurate.
* 0.4
   * More cli output fixes

## License

This project is licensed under the GNU GENERAL PUBLIC LICENSE V3 License - see the LICENSE.md file for details

## TODO
 * Add functionality to catalog and compare the S3 buck with the local manifest
 * Improve error information
 * Add download functionality

## Bugs
 
 * when a file is in the manifest and is not found for upload or splitting, remove it from the database and move on. 