#include <libcc.h>
#include <screen.h>
#include <x86.h>
/*
 * Bit 7	Blinking
 * Bits 6-4	Background color
 * Bit 3	Bright
 * Bit3 2-0	Foreground color
 */

static int csr_x = 0;
static int csr_y = 0;

/*
 * scroll one line
 */
static void
scroll(int lines) {
	int x = MAX_COLUMNS-1, y = MAX_LINES*(lines-1)+MAX_LINES-1;
	short *p = (short *)(VIDEO_RAM+CHAR_OFF(x, y));
	int i = MAX_COLUMNS*(lines-1) + MAX_COLUMNS;
	memcpy((void *)VIDEO_RAM, (void *)(VIDEO_RAM+LINE_RAM*lines),
		   LINE_RAM*(MAX_LINES-lines));
	for (; i>0; --i)
		*p-- = (short)((BLANK_ATTR<<8)|BLANK_CHAR);
}
/*
 * set the cursor position
 */
void
set_cursor(int x, int y) {
	csr_x = x;
	csr_y = y;
	outb(0x3d4,0x0e);
	outb(0x3d5,((csr_x+csr_y*MAX_COLUMNS)>>8)&0xff);
	outb(0x3d4,0x0f);
	outb(0x3d5,((csr_x+csr_y*MAX_COLUMNS))&0xff);
}

void put_c(char c){
    print_c(c,BRIGHT_WHITE,BLACK);
}

void
print_c(char c, COLOUR fg, COLOUR bg) {
	char *p;
	char attr;
	p = (char *)VIDEO_RAM+CHAR_OFF(csr_x, csr_y);
	attr = (char)(bg<<4|fg);
	switch (c) {
	case '\r':
		csr_x = 0;
		break;
	case '\n':
		for (; csr_x<MAX_COLUMNS; ++csr_x) {
			*p++ = BLANK_CHAR;
			*p++ = attr;
		}
		break;
	case '\t':
		c = csr_x+TAB_WIDTH-(csr_x&(TAB_WIDTH-1));
		c = c<MAX_COLUMNS?c:MAX_COLUMNS;
		for (; csr_x<c; ++csr_x) {
			*p++ = BLANK_CHAR;
			*p++ = attr;
		}
		break;
	case '\b':
		if ((! csr_x) && (! csr_y))
			return;
		if (! csr_x) {
			csr_x = MAX_COLUMNS - 1;
			--csr_y;
		} else
			--csr_x;
		((short *)p)[-1] = (short)((BLANK_ATTR<<8)|BLANK_CHAR);
		break;
	default:
		*p++ = c;
		*p++ = attr;
		++csr_x;
		break;
	}
	if (csr_x >= MAX_COLUMNS) {
		csr_x = 0;
		if (csr_y < MAX_LINES-1)
			++csr_y;
		else
			scroll(1);
	}
	set_cursor(csr_x, csr_y);
}

void cga_init(){
    for(int i=0;i<80;i++){
        printf("\n");
    }
    set_cursor(0,0);
}
