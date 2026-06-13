def sum_to(n)
    sum = 0
    i = 1
    while i <= n
        sum += i
        i += 1
    end
    sum
end

n = 10_000_000
result = sum_to(n)
puts result