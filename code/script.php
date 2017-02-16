<?php

$timeStart = microtime(true);

$handle = fopen("resource/NASA_access_log_JulAug95", "r");
if (!$handle) {
    die("Could not open file!\n");
}

$total = 0;
$skipped = 0;
$days = [];
for ($d = 1; $d <= 7; $d++) {
    $hours = [];
    for ($h = 0; $h <= 23; $h++) {
        $hours[$h] = 0;
    }
    $days[$d] = $hours;
}

while ($line = fgets($handle, 1024)) {
    $parts = explode(" ", $line);

    if (sizeof($parts) < 7) {
        die("Unknown format $line.");
    }

    $url = $parts[6];

    // filter only html pages
    if (substr($url, -5) != '.html') {
        $skipped++;
        continue;
    }

    ++$total;

    // extract day in a week and hour in a day

    $date = $parts[3];
    if (substr($date, 0, 1) == '[') {
        $date = substr($date, 1);
    }

    $tz = $parts[4];
    if (substr($tz, -1) == ']') {
        $tz = substr($tz, 0, -1);
    }

    $dt = DateTime::createFromFormat('d/M/Y:H:i:s O', $date . ' ' . $tz);
    if (!$dt) {
        echo "can't parse " . $date . ' ' . $tz . "\n";
        continue;
    }

    $day = $dt->format('N');
    $hour = $dt->format('G');

    // increase visit counter for that day and hour
    $days[$day][$hour]++;
}

fclose($handle);

$max_graph_length = 200;
$graph_scale = 10;
foreach ($days as $d => $hours) {
    switch ($d) {
    case 1:
        echo "Monday\n";
        break;
    case 2:
        echo "Tuesday\n";
        break;
    case 3:
        echo "Wednesday\n";
        break;
    case 4:
        echo "Thursday\n";
        break;
    case 5:
        echo "Friday\n";
        break;
    case 6:
        echo "Saturday\n";
        break;
    case 7:
        echo "Sunday\n";
        break;
    }

    foreach ($hours as $h => $visits) {
        $perc = ($visits / $total) * 100;
        $graph = round($max_graph_length * ($perc / 100) * $graph_scale);

        $n = $h+1;
        if ($n > 23) {
            $n = 0;
        }
        $hour = sprintf("%02d-%02d", $h , $n);

        printf("\t%s\t%.4f\t%0{$graph}s\n", $hour, $perc, "");
    }
}

echo "\nIncluded: $total\n";
echo "Skipped: $skipped\n";

echo "\nElapsed: " . (microtime(true) - $timeStart) . " sec\n";
