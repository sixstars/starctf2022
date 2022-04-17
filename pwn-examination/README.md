# examination

> ret2school once again
>
> solved by 54 teams

This is just an ordinary heap challenge. And the pwn challenge solved by most of the teams.

there exist an integer overflow in paper->score

```c
struct paper
{
    int question_num; // number of questions
    unsigned int score; // down-overflow to get gift:write a bit to leak calloc and heap_addr to overwrite a bit
    char *review; // a pointer to be calloced size 0x20~0x400, calloced
    int review_length;
};
```

The place to trigger integer overflow is in give_score

```c
if(student_case[cnt]->prayed == 1)
{
    puts("the student is lazy! b@d!");
    score-=10; // here causes the interger overflow score will easily be >100
}
```

we can use this to gain the reward, and change tcache idx as well as calloc's mmaped-bit to regain the chunk in the tcache by using calloc.

Then we just  manage to hijack the tcache control chunk to exit_hook. The first one_gadget works here.

> the exp is in ./solve.py

## feedbacks

Most teams thought that the reversing work takes too much time, that's what I actually tried to achieve. Beacuse I found that the real-world-pwn is much more difficult than pwn-challenges in CTFs. So I designed this challenge (And this is the first time that I have ever design a challenge in large CTF games) 

Wish all of you have enjoyed it ! :happy:
