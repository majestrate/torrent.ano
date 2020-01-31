# torrent.ano

Anonymous Torrent Index (with captcha)

## Dependencies

* go 1.9
* postgres

## Building

    $ make

## Setup

    $ cp default.ini config.ini

make sure to edit `config.ini` to have your settings

## Running

    $ ./indextracker config.ini

## Management

adding a new category

    $ ./trackermanager config.ini add-category anime

deleting an existing category

    $ ./trackermanager config.ini del-category anime
    
deleting a torrent

    $ ./trackermanager config.ini del-torrent hexinfohashhere
    
adding a new user for captcha bypass

    $ ./trackermanager config.ini add-user username password

deleting an existing user

    $ ./trackermanager config.ini del-user username


## Demos

anodex:

* [i2p](http://25cb5kixhxm6i6c6wequrhi65mez4duc4l5qk6ictbik3tnxlu6a.b32.i2p/)
* [onion](http://anodex.oniichanylo2tsi4.onion/)

torrent.ano:

* [anonet](http://21.3.37.31/)


## Info

* uploaded torrents are modified by adding an i2p torrent tracker as the default but keeps the rest, for clearnet torrents
* every page has atom feeds, just add `?t=atom` to the url
* http basic auth used to bypass captchas
