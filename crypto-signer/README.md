# signer

Source code is modified from https://github.com/orlp/ed25519 .

The oringal problem is https://github.com/orlp/ed25519/issues/3 .

We remove the patch and add our own patch which doesn't make it any safer: 
    
    r=H(scalar,M)+H(pk[32:],M)


So R will be different when scalar is different. 

But the difference is calculable, so the solution is basic the same:

Sign the same M with two diffrent scalars, then you can calculate the private key. 

Reversing may be the major obstacle in this challenge.


