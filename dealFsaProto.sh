# 删除_get的注释行，目的是为了从description中获取id
start=1
grep -n "_get" fsa178.proto | awk -F ':' '{print $1}' >temp.txt
while read line_num; do
    end=$((line_num - 1));
    sed -n "$start,$end p" fsa178.proto;
    start=$((line_num + 5))
done < temp.txt
sed -n "$start,$ p" fsa178.proto;

