#ifndef _FLAGLIB_H_
#define _FLAGLIB_H_

#include <stdio.h>
#include <stdlib.h>

struct stack {
		char buf[24];
		void *f;
};

const char* get_flag(char* a, int dbg);

#endif
