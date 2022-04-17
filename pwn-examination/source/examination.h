#define MAXSTU 6
extern int fd;
extern unsigned int student_num;
extern struct student* student_case[10];
extern int call_parent_times;
void banner()
{
    setvbuf(stdin,0,2,0);
    setvbuf(stdout,0,2,0);
    fd = open("/dev/urandom",0);
    if(fd == -1)
    {
        puts("open error");
        exit(-1);
    }
    printf(" _____                              _                   _     _                 \n");
    printf("|  ___|                            (_)                 | |   (_)                 \n");
    printf("| |__   __  __   __ _   _ __ ___    _   _ __     __ _  | |_   _    ___    _ __   \n");
    printf("|  __|  \\ \\/ /  / _` | | '_ ` _ \\  | | | '_ \\   / _` | | __| | |  / _ \\  | '_ \\  \n");
    printf("| |___   >  <  | (_| | | | | | | | | | | | | | | (_| | | |_  | | | (_) | | | | | \n");
    printf("\\____/  /_/\\_\\  \\__,_| |_| |_| |_| |_| |_| |_|  \\__,_|  \\__| |_|  \\___/  |_| |_| \n");
}



union process
{
    int unfinished_review;
    long finished_review;
    int checked_review; // cause some partial write
};

union mode
{
    __int8_t pray;
    char* my_mode; // here cause pointer byte rewrite, fixed size 0x20
};



struct paper
{
    int question_num; // number of questions
    unsigned int score; // down-overflow to get gift:write a bit to leak calloc and heap_addr to overwrite a bit
    char *review; // a pointer to be calloced size 0x20~0x400, calloced
    int review_length;
};

struct student
{
    struct paper *test_paper;
    int id;
    union mode cur_mode;
    int prayed;
    int passed;
};


void myread(int fd,char* buf, int len)
{
    read(fd,buf,len);
    int cnt = 0;
    for(cnt = 0;buf[cnt]!='\n';cnt++)
    {
        ;
    }
    buf[--cnt] = '\x00';
    buf[len] = '\xc3';
    return;
}




void menu_teacher()
{
    puts("1. add a student");
    puts("2. give a score");
    puts("3. write a review");
    puts("4. call his/her parent");
    puts("5. change role");
}

void menu_student()
{
    puts("1. do the test");
    puts("2. check for review");
    puts("3. pray");
    puts("4. set mode");
    puts("5. change role");
    puts("6. change id");
}


// void teacher_choose() // show
// {
//     char student_buf[5];
//     int student_chosen = 0;
//     // choose a student number
//     if(student_num == 0)
//     {
//         puts("add some students first!");
//         return;
//     }
//     else // have students
//     {
//         puts("which student id to choose?");
//         read(0,student_buf,5);
//         student_chosen = atoi(student_buf);
//         if(student_chosen<0||student_chosen>9)
//         {
//             puts("please watch carefully :)");
//             return;
//         }
//         struct student *tmp_student = student_case[student_chosen];
//         printf("score: %d\n",tmp_student->test_paper->score);
//         printf("mode: %s\n",*tmp_student->cur_mode.my_mode);
//     }
// }



void add_student()
{
    char student_buf[5];
    int student_chosen = 0;
    // choose a student number
    int question_num=0;
 
    if(student_num>MAXSTU)
    {
        puts("No more students!");
        return;
    }
    struct student *tmp_student = (struct student*)calloc(1,sizeof(struct student));
    struct paper *tmp_paper = (struct paper*)calloc(1,sizeof(struct paper));
    tmp_student->test_paper = tmp_paper;
    student_case[student_num] = tmp_student; // store pointer
    student_num++;
    printf("enter the number of questions: ");
    scanf("%d",&question_num);
    if(question_num>=10 || question_num<=0)
    {
        puts("wrong input!");
        return;
    }
    tmp_student->test_paper->question_num = question_num;
    puts("finish");

}

void give_score()
{
    puts("marking testing papers.....");
    int cnt = 0;
    int score = 0;
    char buf[8];
    for(cnt=0;cnt<student_num;cnt++)
    {
        if(read(fd,buf,8)!=8)
        {
            puts("read_error");
            exit(-1);
        }
        *buf = *buf&0x7f;
        score = (*buf);
        score = score%(student_case[cnt]->test_paper->question_num*10);
        printf("score for the %dth student is %d\n",cnt,score); // if prayed, neg overflow, and a gift for high score
        if(student_case[cnt]->prayed == 1)
        {
            puts("the student is lazy! b@d!");
            score-=10;
        }
        student_case[cnt]->test_paper->score = score;
    }
    puts("finish");
}



void write_review()
{
    int length = 0;
    char choice = 0;
    int cnt = 0;
    printf("which one? > ");
    scanf("%d",&cnt);
    if(student_case[cnt]->test_paper->review!=NULL)
    {
        printf("enter your comment:\n");
        read(0,student_case[cnt]->test_paper->review,student_case[cnt]->test_paper->review_length);
        puts("finish");
        return;
    }
    printf("please input the size of comment: ");
    scanf("%d",&length);
    if(length>=0x400 || length<=0)
    {
        puts("wrong length :'(");
        return ;
    }
    student_case[cnt]->test_paper->review = calloc(1,length);
    printf("enter your comment:\n");
    read(0,student_case[cnt]->test_paper->review,length);
    student_case[cnt]->test_paper->review_length = length;


    puts("finish");
}

void call_parent()
{
    char student_buf[10];
    int student_chosen = 0;
    puts("only 3 chances to call parents!");
    if(call_parent_times == 0)
    {
        puts("no you can't");
        return;
    }
    else
    {
        --call_parent_times;
        if(student_num == 0)
        {
            puts("add some students first!");
            return;
        }
        else // have students
        {
            puts("which student id to choose?");
            read(0,student_buf,5);
            student_chosen = atoi(student_buf);
            if(student_chosen<0||student_chosen>9|| student_case[student_chosen] == NULL)
            {
                puts("please watch carefully :)");
                return;
            }
            printf("bad luck for student %d! Say goodbye to him/her!",student_chosen);
            if(student_case[student_chosen]->test_paper->review!=NULL)
            {
                free(student_case[student_chosen]->test_paper->review);
            }
            free(student_case[student_chosen]->test_paper);
            free(student_case[student_chosen]);
            student_case[student_chosen]=0;
            student_num--;
        }
    }

}

void do_test(int id)
{
    sleep(1);
    puts("ok, finished");
}

int exitread(int fd,char* buf,int len)
{
    int tmp1;
    while(len--)
    {
        tmp1 = read(fd,buf,1);
        if(!tmp1 || tmp1 == -1)
        return 0;
        else
        {
            if(*buf == '\n')
            {
                return 1;
            }
            ++buf;
        }
    }
    return 0;
}


void quit()
{
    puts("never pray again!");
    char* buf = (char*)malloc(0x300);
    exitread(0,buf,0x300);
    exit(-1);
}



void setmode(int id)
{
    
    if(student_case[id]->prayed == 1)
    {
        int tmp;
        __int8_t score;
        puts("enter your pray score: 0 to 100");
        scanf("%d",&tmp); // one byte overwrite the least byte of heap_addr of oneself(because of union)
        if(tmp<0 || tmp >100)
        {
            puts("bad!");
            return;
        }
        score = tmp;
        student_case[id]->cur_mode.pray = score; // change to tcache, and write to corresponding fastbin
    }
    else // no pray
    {
        if(!student_case[id]->cur_mode.my_mode)
        {
            student_case[id]->cur_mode.my_mode = calloc(1,0x20);
        }
        puts("enter your mode!");
        read(0,student_case[id]->cur_mode.my_mode,0x20);
    }
    puts("finish");
}

void check(int id) // check teacher's review
{
    if(student_case[id]->passed == 1)
    {
        puts("already gained the reward!");
        return;
    }
    char* addr;
    char addrbuf[0x10];
    if(student_case[id]->test_paper->score>=90)
    {
        printf("Good Job! Here is your reward! %p\n",student_case[id]);
        printf("add 1 to wherever you want! addr: ");
        myread(0,addrbuf,0x10);
        addr = (char*)atol(addrbuf);
        *addr+=1; // change calloc mmaped bit to trigger leak of libc 
        student_case[id]->passed = 1;
    }
    if(student_case[id]->test_paper->review!=NULL)
    {
        puts("here is the review:");
        write(1,student_case[id]->test_paper->review,student_case[id]->test_paper->review_length);
    }
    else
    {
        puts("no reviewing yet!");
        return ;
    }
}

void pray(int id)
{
    puts("prayer...Good luck to you");
    student_case[id]->prayed^=1;
    puts("finish");
}


int change_role()
{
    int tmp_role;
    printf("role: <0.teacher/1.student>: ");
    scanf("%d",&tmp_role);
    return tmp_role;
}

int change_id()
{
    int tmp_id;
    printf("input your id: ");
    scanf("%d",&tmp_id);
    
    if(student_case[tmp_id] == NULL || tmp_id>MAXSTU)
    {
        puts("RUA alien?");
        exit(-1);
    }
    printf("hello, student %d\n",tmp_id);
    return tmp_id;
}