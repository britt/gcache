# github.com/britt/GCache

**Status: functional but undocumented aside form comments**

A fork of [Jun Kimura's GCache](https://github.com/bluele/gcache) that adds the ScoreCache type. 
ScoreCache uses a scoring function to identify candidates for eviction. Lowest
scored items are evicted first. It also uses an aggregate metric rather (like total bytes cached) 
than a number of items to determine when evictions are necessary.

For documentation see [the original GCache](https://github.com/bluele/gcache).
