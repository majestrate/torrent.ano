# torrent.ano

Anonymous Torrent Index (with captcha)

## Dependencies

* go 1.8
* postgres

## Building

    $ make

## Running

    $ ./indextracker

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
