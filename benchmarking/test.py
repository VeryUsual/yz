def sum_to(n):
    sum = 0
    i = 1
    while i <= n:
        sum += i
        i += 1
    return sum

n = 10000000
result = sum_to(n)
print(result)