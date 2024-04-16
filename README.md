My translation of the code from the [Hypermedia Systems](https://hypermedia.systems/) book.

## Notes
1. I decied to skip translating the flask.Flash() function. While nice, I deemed it a distraction from my goal of learnign htmx.
2. The book implemented three forms of data partitioning(?) in the table. Paging, Click To Load, and Infinite Scrolling. Paging, Click To Load, and Infinite Scrolling have
thier own branches. Infinite scrolling was last implemented and it was messing with the active search, and I'm not sure how to make them work togehter. Plus infinite scroll is funky in this
contacts example.
3. The book code has a random sleep time in the archiver, but the random part doesn't seem necessary. It wasn't working and I didn't want to sepnd time working on it so I switched to using a second and it works fine.

## Feature Ideas
1. Update contact count in index.html when a contact is deleted using the inline delete.
2. Fix styling cause it kinda sucks (but good styling is not the goal here!)


