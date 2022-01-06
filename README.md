# torrent.ano

Anonymous Torrent Index (with captcha)

## Dependencies

* go
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

## Info

* uploaded torrents are modified by adding an i2p torrent tracker as the default but keeps the rest, for clearnet torrents
* every page has atom feeds, just add `?t=atom` to the url
* http basic auth used to bypass captchas
