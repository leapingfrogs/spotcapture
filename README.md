# SpotCapture

This is a trivial command line app that supports adding/removing the currently playing Spotify track to/from a playlist.

## Usage:

### 1. Add current track to playlist:
```bash
> spotcapture
```

### 2. Remove current track from playlist:
```bash
> spotcapture -remove
```

## Notes

The first time you run spotcapture it will launch a browser window prompting you to authenticate with spotify.  The generated token will be stored in ``~/.spotcapture`.  Additionally a private playlist will be created in your spotify account named `SpotCatpture`, the id of this playlist is also stored in ``~/.spotcapture` along with your spotify user id.
If your user token expires you will be prompted to authenticate again.