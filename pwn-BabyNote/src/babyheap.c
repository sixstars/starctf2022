#include <stdio.h>
#include <string.h>
#include <stdint.h>
#include <stdlib.h>
#include <unistd.h>

static void banner()
{
    puts("                                                   ");
    puts("    _/  _/  _/        _/_/_/  _/_/_/_/_/  _/_/_/_/ ");
    puts("     _/_/_/        _/            _/      _/        ");
    puts("  _/_/_/_/_/      _/            _/      _/_/_/     ");
    puts("   _/_/_/        _/            _/      _/          ");
    puts("_/  _/  _/        _/_/_/      _/      _/           ");
    puts("                                                   ");
    puts("                                                   ");
}

static int recvuntil(void *buf, size_t n)
{
    for (int i = 0; i < n; i++)
    {
        char c;
        if (read(0, &c, 1) != 1)
        {
            return i;
        }
        ((char *)buf)[i] = c;
        if (c == '\n')
        {
            ((char *)buf)[i] = 0;
            return i;
        }
    }
    return n;
}

static int readint()
{
    char buf[0x10] = {0};
    recvuntil(&buf, sizeof(buf));
    return atoi(buf);
}

static size_t read_name(uint8_t **name)
{
    printf("name size: ");
    size_t name_size = readint();
    *name = calloc(1, name_size);
    printf("name: ");
    return recvuntil(*name, name_size);
}

static size_t read_note(uint8_t **note)
{
    printf("note size: ");
    size_t note_size = readint();
    *note = calloc(1, note_size);
    printf("note content: ");
    return recvuntil(*note, note_size);
}

struct node
{
    uint8_t *name;
    uint8_t *note;
    size_t name_size;
    size_t note_size;
    struct node *next;
};

static void value_dump(const uint8_t *data, size_t size)
{
    printf("%#lx:", size);
    for (int i = 0; i < size; i++)
    {
        printf("%02x", data[i]);
    }
    puts("");
}

static struct node *list_head = NULL;

static void menu()
{
    puts("--------menu-------");
    puts("1: add a note");
    puts("2: find a note");
    puts("3: delete a note");
    puts("4: forget all notes");
    puts("5: exit");
    printf("option: ");
}

static struct node *lookup(const uint8_t *name, size_t name_size)
{
    for (struct node *n = list_head; n; n = n->next)
    {
        if (n->name_size == name_size && !memcmp(name, n->name, name_size))
        {
            return n;
        }
    }
    return NULL;
}

static void addNote()
{
    struct node *node = calloc(1, sizeof(struct node));
    node->name_size = read_name(&node->name);
    // always insert to the head, don't check duplicated entries
    node->note_size = read_note(&node->note);
    node->next = list_head;
    list_head = node;
    puts("ok");
}

static void findNote()
{
    uint8_t *name = NULL;
    size_t name_size = read_name(&name);
    struct node *n = lookup(name, name_size);
    if (n == NULL)
    {
        puts("oops.....");
    }
    else
    {
        value_dump(n->note, n->note_size);
    }
    free(name);
}

static void deleteNote()
{
    uint8_t *name = NULL;
    size_t name_size = read_name(&name);
    struct node *n = lookup(name, name_size);
    if (n == NULL)
    {
        puts("oops.....");
    }
    else
    {
        if (list_head == n && list_head->next == NULL)
        {
            list_head = NULL;
        }
        else if (n->next != NULL)
        {
            struct node **p = &list_head;
            while (*p != n)
            {
                p = &(*p)->next;
            }
            *p = n->next;
        }
        free(n->name);
        free(n->note);
        free(n);
        puts("ok");
    }
    free(name);
}


static void forgetNote()
{
    list_head = NULL;
}

int main(int argc, char **argv)
{
    setvbuf(stdout, NULL, _IONBF, 0);
    setvbuf(stderr, NULL, _IONBF, 0);
    banner();
    while (1)
    {
        menu();
        int op = readint();
        switch (op)
        {
        case 1:
            addNote();
            break;
        case 2:
            findNote();
            break;
        case 3:
            deleteNote();
            break;
        case 4:
            forgetNote();
            break;
        case 5:
            puts("bye");
            exit(0);
        default:
            puts("invalid");
            exit(0);
        }
    }
    return 0;
}
