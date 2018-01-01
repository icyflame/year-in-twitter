users=(CaseyNeistat icyflame ghostwriternr beenharsh SemiColonComma realdonaldtrump kendalljenner taylorswift13 lovebillynyc AyushPandey9 kritarthjha _icyflame HilaKleinH3 h3h3productions)
for i in ${users[*]}; do
    echo "Getting $i"
    /usr/bin/time -f "%E" curl -s "https://twitter.siddharthkannan.in/get/$i" > /dev/null
done;
