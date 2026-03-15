export interface TreeNode {
  name: string
  path: string
  isFolder: boolean
  children: TreeNode[]
}

export function buildTree(paths: string[]): TreeNode {
  const root: TreeNode = { name: "", path: "", isFolder: true, children: [] }

  const sortedPaths = [...paths].sort()

  for (const path of sortedPaths) {
    const segments = path.split("/").filter(Boolean)
    let current = root

    for (let i = 0; i < segments.length; i++) {
      const segment = segments[i]
      const isLast = i === segments.length - 1
      const currentPath = "/" + segments.slice(0, i + 1).join("/")

      let child = current.children.find((c) => c.name === segment)

      if (!child) {
        child = {
          name: segment,
          path: currentPath,
          isFolder: !isLast || paths.some((p) => p.startsWith(currentPath + "/")),
          children: [],
        }
        current.children.push(child)
      }

      current = child
    }
  }

  sortTree(root)
  return root
}

function sortTree(node: TreeNode): void {
  node.children.sort((a, b) => {
    if (a.isFolder !== b.isFolder) {
      return a.isFolder ? -1 : 1
    }
    return a.name.localeCompare(b.name)
  })

  for (const child of node.children) {
    sortTree(child)
  }
}

export function findNode(root: TreeNode, path: string): TreeNode | null {
  if (root.path === path) return root

  for (const child of root.children) {
    const found = findNode(child, path)
    if (found) return found
  }

  return null
}
