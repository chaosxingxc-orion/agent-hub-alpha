#!/bin/bash
# reapply_skill_preview.sh — Trae 重建 w06-checker 容器后，一键恢复 skill 内容预览能力
# 恢复：① server 的 /v1/skills/{id}/content 端点 ② 前端 skill 详情内容预览区 ③ /skills 技能文件
# 用法：bash ~/reapply_skill_preview.sh
set -e
C=w06-checker
STAGE=/tmp/skill-preview-stage
rm -rf $STAGE && mkdir -p $STAGE
ASSETS=~/w06-skill-assets

echo "[1/6] 恢复 server（直拷预置补丁版：content+files 端点）"
if [ -f $ASSETS/CheckerHttpServer.patched.java ]; then
  docker cp $ASSETS/CheckerHttpServer.patched.java $C:/src/main/java/com/agenthub/w06/checker/server/CheckerHttpServer.java
  echo "  已恢复（$(wc -c < $ASSETS/CheckerHttpServer.patched.java) 字节）"
else
  echo "  警告：$ASSETS/CheckerHttpServer.patched.java 不存在"
fi

echo "[3/6] 部署 skill 文件到容器"
if [ -d $ASSETS/skills ]; then
  docker exec $C rm -rf /skills 2>/dev/null || true
  docker cp $ASSETS/skills $C:/skills
  echo "  已从 $ASSETS/skills 恢复"
else
  echo "  警告：$ASSETS/skills 不存在"
fi

echo "[4/6] 恢复前端 workbench.html（含内容预览区）"
if [ -f $ASSETS/workbench.html ]; then
  if docker exec $C grep -q "renderMarkdown" /app/workbench.html 2>/dev/null; then
    echo "  前端已含预览区，跳过"
  else
    docker cp $ASSETS/workbench.html $C:/app/workbench.html
    echo "  前端预览区已恢复"
  fi
else
  echo "  警告：$ASSETS/workbench.html 不存在（预览区无法恢复）"
fi

echo "[5/6] 容器内重编译"
docker exec $C bash -c 'cd /src && CP=$(ls libs/*.jar | tr "\n" ":") && javac -cp $CP -d out/main $(find main -name "*.java")' && echo "  编译 OK"

echo "[6/6] 重启容器"
docker restart $C >/dev/null && sleep 10

echo "验证（应为 JSON 含 content 字段）："
curl -s http://127.0.0.1:18083/v1/skills/annual-report/content | head -c 200
echo
echo "完成。注意：若 Trae 工程后续更新了 workbench.html，需重新生成 $ASSETS/workbench.html（含预览区版本）。"
