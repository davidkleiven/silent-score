![image](logo.svg)


# Silent Score

Silent Score is an application that simplifies the process of compiling a score for a movie. It is shipped with a library containing a selection of silent movie-era songs, but it can also use songs from a custom folder. The program reads `.musicxml` and `.mxl` (compressed MusicXML) files.

The application features a terminal user interface, and a help section displaying available key combinations within a view is shown at the bottom of the screen.

The program generates a compiled score by selecting pieces from the standard library and any local folder the user has configured. In the project workspace, the user populates the following table:

| Scene Description | Tempo | Keywords | Theme | Duration |
| ----------------- | ----- | -------- | ----- | -------- |
| Description of the scene that will appear as *Staff text* | Tempo (beats per minute) of the piece (if not specified, the tempo is extracted from the chosen score) | Text describing the type of music desired. Any text field within a `.musicxml` or `.mxl` file is used for matching. Examples may be composer, agitato, allegro, waltz, foxtrott etc. The piece with text that has the highest similarity with the text in the keyword field will be selected for the scene | Scenes with the same theme number are guaranteed to use the same piece | Duration of the scene in seconds |
