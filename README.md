# Document generator for RBLR1000

I produce rider documentation for the RBLR1000 event run by IBAUK on behalf of the RBLR.

To do this I read a database of rider details held in ScoreMaster or Alys format and a series of templates for certificates, 
receipt logs, disclaimers, etc customised by route. 

The routes are:- A-NCW [2], B-NAC [1], C-SCW [4], D-SAC [3], E-5CW [6], F-5AC [7]. The numbers in brackets are the 'class' field
in the ScoreMaster db, the tags are held in the 'Route' field of an Alys db representing each route.

Each document type is held in its own subfolder. At least three files must exist for each document: header.html, footer.html contain
ordinary html to be included at the start and end of the output file. Each entrant's output is formatted using a template file
either entrant.html or entrant*n*.html where *n* is the Class of the entrant.

## Certificates
Certificates for the RBLR1000 are always printed using this utility. Certificates for other rallies are printed directly from ScoreMaster. This will also print the post-event certificates for people who switched North/South routes or finished after 24 hours. Use **-final** to print those.

## Receipt logs
Customised receipt logs are printed for each rider including the essential receipt points on the various designated routes.

## Disclaimers
Disclaimers are produced for the RBLR1000 and for other UK rallies.

## Handbook
The subfolder *handbook* contains a plaintext HTML version of the "detailed instructions" used on the website. It can be edited using any plaintext editor and converted to PDF using a browser's print facility. 