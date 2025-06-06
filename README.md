# CleanSync
```
   ________                _____                 
  / ____/ /__  ____ _____ / ___/__  ______  _____
 / /   / / _ \/ __ `/ __ \\__ \/ / / / __ \/ ___/
/ /___/ /  __/ /_/ / / / /__/ / /_/ / / / / /__  
\____/_/\___/\__,_/_/ /_/____/\__, /_/ /_/\___/  
                             /____/              
```


## Description

A little personal project I use to keep my videos backed up to S3. I've spent many hours processing my video library so I thought is was worth a bit of effort and a few buck a month to upload them to the S3 Glacier Deep Archive in the case that anything happens to my local media.

## Getting Started

Download and build with Go 1.21. Uses sqlite so you will need the gcc libries to build. Download GCC from (http://tdm-gcc.tdragon.net/download)[http://tdm-gcc.tdragon.net/download]

### Dependencies

Uses sqlite so you will need the gcc libries to build. Download GCC from (http://tdm-gcc.tdragon.net/download)[http://tdm-gcc.tdragon.net/download]
Was built targeting windows 10, but I don't see any reason it would not work on Linux or Mac

*Requires ffmpeg to me installed and in your path*

### Installing

After it is built, copy it to where you want it to live. It will create a manifest.db in that directory when ran. This is where it catalogs the files that it backs up.

### Executing program


```
cleansync -help # for information

```

```
NAME:
   cleansync - tools to manage video library

USAGE:
   cleansync [global options] command [command options]

COMMANDS:
   adclear  Removes adds from the source and copies the resulting video to the destination
   sync     upload new files to the provided bucket
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help
```

*Subcommands* 

* sync
  * `.\cleansync.exe sync -path=x:\videos -bucket=my-backup-bucket -filter=mkv -filter=mp4 -deep`
                                                             
```
NAME:
   cleansync sync - upload new files to the provided bucket

USAGE:
   cleansync sync [command options]

OPTIONS:
   --path value, -p value                                 The source (local) folder to sync with S3
   --bucket value, -b value                               The name of the bucket to sysnc to
   --filter value, -f value [ --filter value, -f value ]  file types to filter for. Can be specified multiple times for multiple file types.
   --deep, -d                                             deep archive in S3 (default: false)
   --help, -h                                             show help                                         show help
```

* adclear
  * `./cleansync adclear --source c:\artifacts\original.mp4 --dest x:\artifacts\edited3.mp4 --skip_first`

NAME:
   cleansync adclear - Removes adds from the source and copies the resulting video to the destination

USAGE:
   cleansync adclear [command options]

OPTIONS:
   --source value  The source file or folder, if it is a folder, it will attempt to process all video files. (currently mp4, mkv)
   --dest value    The destination file or folder
   --skip_first    Skips the first chapter, thus omiting it from the final product. Usefull for removing that 'Recorded by...' at the begining of playon videos (default: false)
   --help, -h      show help

## Version History

* 0.0.1
    * Initial Release
* v0.0.2
    * Added file splitting for files greater than 4GB
* v0.0.3
   * Cleaned up cli output. Now it is simpler, but accurate.
* v0.0.4 
   * started tagging versions
* v0.0.5
   * More cli output fixes
*v0.0.6
   * Fixed the crash if there are no commercials detected
   * Fix hang when there is a quote in the file name. Renames the file without the quote.
   * Made the file transfer directly after the commercials are removed, instead of after all the videos are processed
   * various code clean up.

## License

This project is licensed under the GNU GENERAL PUBLIC LICENSE V3 License - see the LICENSE.md file for details

## TODO
 * Add functionality to catalog and compare the S3 bucket with the local manifest
 * Improve error information
 * Add download functionality
 * Change UI to https://github.com/charmbracelet/bubbletea/tree/master/examples
 * Need to confirm that the destiation is a directory and not a file when trying to copy files to thier dest during the adclear command

## Bugs
 
 * when a file is in the manifest and is not found for upload or splitting, remove it from the database and move on. not
 * need to add overall status to cli output

 ## Notes

 In mid process in the conversion to charmbracelet. I think I have the concepts worked out. I need to finish working through the upload section of the logic.
 Working on fixing live updates to the progress bar for splitting files up

 Integrate "active" properties on the progress readers to select when a progress bar update should be active.

 Added commercial removal subcommand - goal is to remove the commercials from a specified directory and upload to media server with one command. Right now can only specify on video at a time
