<?php

function sum_to($n) {
    $sum = 0;
    $i = 1;
    while ($i <= $n) {
        $sum += $i;
        $i += 1;
    }
    return $sum;
}

$N = 10000000;
$result = sum_to($N);
echo "$result\n";

?>