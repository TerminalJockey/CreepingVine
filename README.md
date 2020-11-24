# CreepingVine
____
simple crossplatform go c2/botnet with mongodb backend.<strike> Not gonna provide support or help setting up.
its pretty straight forward and if you cant figure it out you probably shouldnt have a botnet.
Just improving my skills here, </strike> this wont evade detections or work particularly well (yet) so dont use it for evil (ever).

update!
___
I decided I like this project and will be actively developing this further.

My plans for this project:
I am going for a resilient and distributed C2 written in as close to pure Go as possible. Ideally, this will be a modular
system in which we can deploy new plugins, although go poses some hurdles to this implementation. I also like the idea of
being able to promote nodes to an "administrator" or "orchestrator" role so we can migrate our admin panels around the net
and evade capture. Current plan is to implement a few c2 channels, ideally dns tunneling, http/https comms, and probably a
straight tcp connection, might play around with a custom p2p deal as well.

Imagine ivy, creeping up a brick wall, finding foodholds and being a real pain in the ass to kill. 
