// gcc -s ./examination.c -o exam_release
#include <stdio.h>
#include <malloc.h>
#include <string.h>
#include <stdlib.h>
#include <unistd.h>
#include <time.h>
#include <sys/types.h>    
#include <sys/stat.h>    
#include <fcntl.h>
#include"examination.h"

int role = -1; // role=0:teacher,  role=1: student

const int teacher_num=1;
unsigned int student_num=0;
int lock = 0;
struct student* student_case[10];
int fd;
int call_parent_times = 4;





int main()
{
    banner();
    char role_buf[10];
    char choice_buf[5];
    int choice_teacher;
    int choice_student;
    printf("role: <0.teacher/1.student>: ");
    scanf("%d",&role);
    

    while(1)
    {
        
        if(role == 0) // teacher
        {
            menu_teacher();
            printf("choice>> ");
            read(0,choice_buf,2);
            choice_teacher = atoi(choice_buf);
            switch(choice_teacher)
            {
                case 1:
                add_student();
                break;
                case 2:
                give_score();
                break;
                case 3:
                write_review();
                break;
                case 4:
                call_parent();
                break;
                case 5:
                role = change_role();
                break;
                case 6:
                quit();
                break;
            }
        }
        else // student
        {
            int id = 0;
            if(student_num == 0)
            {
                puts("no student yet");
                break;
            }
            while(role == 1)
            { 
                menu_student();
                printf("choice>> ");
                read(0,choice_buf,2);
                choice_student = atoi(choice_buf);
                switch(choice_student)
                {
                    case 1: 
                    do_test(id);
                    break;
                    case 2:
                    check(id);
                    break;
                    case 3:
                    pray(id);
                    break;
                    case 4:
                    setmode(id);
                    break;
                    case 5:
                    role = change_role();
                    break;
                    case 6:
                    id = change_id();
                    break;
                }
            }
        }
    }
    

}
