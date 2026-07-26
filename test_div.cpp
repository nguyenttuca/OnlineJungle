#include <iostream>
int main() {
    volatile int a = 0;
    std::cout << 1 / a;
    return 0;
}
