#!/bin/bash
#
# multica-upgrade.sh — 三方更新脚本
# 功能：拉取原版 multica 并自动合并 Gitee 相关修改
#
set -e

MULTICA_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$MULTICA_DIR"

UPSTREAM_REMOTE="upstream"
UPSTREAM_URL="https://github.com/multica-ai/multica.git"
ORIGIN_URL=$(git remote get-url origin 2>/dev/null || echo "")

echo "=== Multica 三方更新 ==="
echo ""

# 1. 检查 upstream 远端
echo "[1/6] 检查 upstream 远端..."
if ! git remote get-url "$UPSTREAM_REMOTE" &>/dev/null 2>&1; then
    echo "  添加 upstream 远端: $UPSTREAM_URL"
    git remote add "$UPSTREAM_REMOTE" "$UPSTREAM_URL"
else
    echo "  upstream 远端已存在"
fi

# 2. 获取当前版本
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || git log --oneline -1 | cut -d' ' -f1)
echo "  当前版本: $CURRENT_VERSION"

# 3. Fetch upstream 最新代码
echo ""
echo "[2/6] 获取 upstream 最新代码..."
git fetch "$UPSTREAM_REMOTE" --tags --force

# 4. 获取 upstream 最新版本
UPSTREAM_LATEST=$(git describe --tags "${UPSTREAM_REMOTE}/main" 2>/dev/null || echo "最新")
echo "  upstream 最新版本: $UPSTREAM_LATEST"

# 5. 检查是否有 Gitee 相关修改
echo ""
echo "[3/6] 检查 Gitee 相关修改..."
GITEE_FILES=(
    "server/internal/vcs/gitee"
    "packages/views/settings/components/gitee-tab.tsx"
    "packages/core/types/gitee.ts"
    "packages/core/gitee"
)

HAS_GITEE_CHANGES=false
for file in "${GITEE_FILES[@]}"; do
    if [ -e "$file" ] || git log --all --oneline | grep -q "gitee"; then
        HAS_GITEE_CHANGES=true
        echo "  发现 Gitee 相关文件: $file"
    fi
done

if [ "$HAS_GITEE_CHANGES" = false ]; then
    echo "  未发现 Gitee 相关修改，继续标准更新流程"
fi

# 6. 暂存本地修改
echo ""
echo "[4/6] 暂存本地修改..."
if git stash list | grep -q .; then
    echo "  已有 stash 记录，保存到新 stash"
fi
git stash push -m "multica-upgrade-pre-$(date +%Y%m%d%H%M%S)" --keep-index

# 7. 合并 upstream/main
echo ""
echo "[5/6] 合并 upstream/main..."
MERGE_RESULT=$(git merge "${UPSTREAM_REMOTE}/main" --no-edit 2>&1 || echo "MERGE_FAILED")
if echo "$MERGE_RESULT" | grep -q "MERGE_FAILED\|CONFLICT"; then
    echo "  检测到合并冲突！"
    echo "  冲突文件:"
    git diff --name-only --diff-filter=U | while read file; do
        echo "    - $file"
    done
    echo ""
    echo "  请手动解决冲突后运行: git add . && git commit -m 'merge upstream'"
    echo "  或运行 git merge --abort 取消合并"
    exit 1
fi
echo "  合并成功"

# 8. 恢复 Gitee 相关文件
echo ""
echo "[6/6] 检查并恢复 Gitee 相关修改..."
# 尝试从 stash 恢复 Gitee 文件（如果之前有 Gitee 文件的话）
if git stash list | grep -q "multica-upgrade-pre"; then
    STASH_LIST=$(git stash list --format="%s" | grep "multica-upgrade-pre" | head -1 | cut -d: -f2 | tr -d ' ')
    if [ -n "$STASH_LIST" ]; then
        # 检查 stash 中是否有 Gitee 相关文件
        for file in "${GITEE_FILES[@]}"; do
            if git stash show "stash@{0}" --name-only 2>/dev/null | grep -q "$file"; then
                echo "  恢复 Gitee 文件: $file"
                git checkout "stash@{0}" -- "$file" 2>/dev/null || true
            fi
        done
    fi
fi

# 9. 输出结果
echo ""
echo "=== 更新完成 ==="
echo "  当前版本: $(git describe --tags --abbrev=0 2>/dev/null || git log --oneline -1 | cut -d' ' -f1)"
echo ""
echo "如果有任何冲突，请手动解决后提交"
echo "运行 'pnpm build && make build' 验证编译"