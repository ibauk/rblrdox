# Document generator for RBLR1000

I produce rider documentation for the RBLR1000 event run by IBAUK on behalf of the RBLR.

To do this I read a database of rider details held in ScoreMaster format and a series of templates for certificates, 
receipt logs, disclaimers, etc customised by route. 

The routes are:- A-NC [2], B-NAC [1], C-SC [4], D-SAC [3], E-5C [6], F-5AC [7]. The numbers in brackets are the 'class' field
in the ScoreMaster db representing each route.

Each document type is held in its own subfolder. At least three files must exist for each document: header.html, footer.html contain
straight html to be included at the start and end of the output file. Each entrant's output is formatted using a template file
either entrant.html or entrant<n>.html where <n> is the Class of the entrant.

## Certificates
Although it is possible to produce certificates using this package, it is probably better to use the certificate function built into
ScoreMaster itself as that version is periodically improved.

## Handbook
The subfolder *handbook* contains a plaintext HTML version of the "detailed instructions" used on the website. It can be edited using any plaintext editor and converted to PDF using a browser's print facility. 