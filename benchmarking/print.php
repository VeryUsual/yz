<?php

echo 1 + 1 . "\n";

$y = 1;
$thesum = 0;

while ($y < 1000) {
    $thesum += $y;
    $y += 1;
}

echo $thesum . "\n";

?>