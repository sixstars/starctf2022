## treasure-hunters

本题是rwctf 4th treasure-hunter的拓展，其目的是让用户体会到SMT库设计的优点，主要漏洞是merge函数是合并并没有考虑左右子树的顺序，具体漏洞思路参考 https://ctftime.org/writeup/32181

本题在这之上加了一个team 需要至少四个人一起才能完成任务拿到flag，所以攻击者需要重新构造proof 达到目的，这里面主要是希望选手使用0x48这个操作码，利用思路和之前题目一摸一样。

This question is based on rwctf 4th treasure-hunter, its purpose is to allow users to appreciate the advantages of SMT library design, the main vulnerability is that the merge function does not take into account the order of the left and right subtree. For specific exploitation refer to https://ctftime.org/writeup/32181.

This question added a TEAM on top of this, need at least four people together to complete the task to get the flag, so the attacker needs to reconstruct the PROOF to achieve the purpose. It is supposed that the player use 0x48 this opcode, using the same idea with the previous challenge.
