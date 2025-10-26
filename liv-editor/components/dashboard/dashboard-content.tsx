"use client"

import { useState, useEffect } from "react"
import Link from "next/link"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Plus, FileText, Clock, Settings, HelpCircle, Search, Trash2, Share2 } from "lucide-react"
import { useButtonActions } from "@/hooks/use-button-actions"

const recentProjects = [
  {
    id: 1,
    name: "Q4 Product Roadmap",
    date: "2 days ago",
    thumbnail: "bg-gradient-to-br from-primary to-accent",
    size: "2.4 MB",
  },
  {
    id: 2,
    name: "Design System v2",
    date: "1 week ago",
    thumbnail: "bg-gradient-to-br from-accent to-secondary",
    size: "1.8 MB",
  },
  {
    id: 3,
    name: "Marketing Deck",
    date: "2 weeks ago",
    thumbnail: "bg-gradient-to-br from-secondary to-primary",
    size: "3.2 MB",
  },
  {
    id: 4,
    name: "Annual Report",
    date: "3 weeks ago",
    thumbnail: "bg-gradient-to-br from-primary/80 to-accent/80",
    size: "4.1 MB",
  },
  {
    id: 5,
    name: "Brand Guidelines",
    date: "1 month ago",
    thumbnail: "bg-gradient-to-br from-accent/80 to-secondary/80",
    size: "2.9 MB",
  },
  {
    id: 6,
    name: "Pitch Deck",
    date: "1 month ago",
    thumbnail: "bg-gradient-to-br from-secondary/80 to-primary/80",
    size: "3.5 MB",
  },
]

const templates = [
  {
    id: 1,
    name: "Blank Document",
    icon: (
      <svg className="w-8 h-8" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
        <polyline points="14 2 14 8 20 8" />
      </svg>
    ),
  },
  {
    id: 2,
    name: "Presentation",
    icon: (
      <svg className="w-8 h-8" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <rect x="2" y="7" width="20" height="15" rx="2" ry="2" />
        <polyline points="17 2 12 7 7 2" />
      </svg>
    ),
  },
  {
    id: 3,
    name: "Report",
    icon: (
      <svg className="w-8 h-8" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
        <line x1="9" y1="9" x2="15" y2="9" />
        <line x1="9" y1="13" x2="15" y2="13" />
        <line x1="9" y1="17" x2="13" y2="17" />
      </svg>
    ),
  },
  {
    id: 4,
    name: "Portfolio",
    icon: (
      <svg className="w-8 h-8" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <rect x="2" y="7" width="20" height="14" rx="2" ry="2" />
        <path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16" />
      </svg>
    ),
  },
]

export function DashboardContent() {
  const [searchQuery, setSearchQuery] = useState("")
  const [animateIn, setAnimateIn] = useState(false)
  const [deletingId, setDeletingId] = useState<number | null>(null)
  const { executeAction } = useButtonActions()

  useEffect(() => {
    setAnimateIn(true)
  }, [])

  const filteredProjects = recentProjects.filter((p) => p.name.toLowerCase().includes(searchQuery.toLowerCase()))

  const handleDeleteProject = async (projectId: number) => {
    await executeAction(async () => {
      setDeletingId(projectId)
      await new Promise((resolve) => setTimeout(resolve, 600))
      setDeletingId(null)
    }, `delete-${projectId}`)
  }

  const handleShareProject = async (projectName: string) => {
    await executeAction(async () => {
      await navigator.clipboard.writeText(`Check out my project: ${projectName}`)
    }, `share-${projectName}`)
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b border-border bg-gradient-to-r from-background to-muted/50 backdrop-blur-sm sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-primary to-accent flex items-center justify-center glow-effect">
              <span className="text-primary-foreground font-bold text-lg">L</span>
            </div>
            <h1 className="text-2xl font-bold">LIV Editor</h1>
          </div>
          <div className="flex items-center gap-4">
            <button className="text-muted-foreground hover:text-accent transition-colors hover:scale-110 duration-200">
              <HelpCircle size={20} />
            </button>
            <Link href="/settings">
              <button className="text-muted-foreground hover:text-accent transition-colors hover:scale-110 duration-200">
                <Settings size={20} />
              </button>
            </Link>
            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary to-accent cursor-pointer hover:opacity-80 transition-opacity glow-effect" />
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-6 py-12">
        {/* Hero Section */}
        <div
          className={`mb-16 transition-all duration-700 ${
            animateIn ? "opacity-100 translate-y-0" : "opacity-0 translate-y-4"
          }`}
        >
          <h2 className="text-4xl font-bold mb-4 bg-gradient-to-r from-primary to-accent bg-clip-text text-transparent">
            Welcome back
          </h2>
          <p className="text-muted-foreground text-lg mb-8">
            Create, edit, and share interactive documents with live animations and embedded media.
          </p>

          <div className="flex gap-4 flex-wrap">
            <Link href="/editor">
              <Button
                size="lg"
                className="gap-2 bg-gradient-to-r from-primary to-accent hover:shadow-lg hover:shadow-primary/50 transition-all glow-effect"
              >
                <Plus size={20} />
                New Document
              </Button>
            </Link>
            <Button
              size="lg"
              variant="outline"
              className="hover:shadow-lg transition-all hover:bg-accent/20 hover:text-accent bg-transparent"
            >
              Open File
            </Button>
          </div>
        </div>

        {/* Templates Section */}
        <div
          className={`mb-16 transition-all duration-700 delay-100 ${
            animateIn ? "opacity-100 translate-y-0" : "opacity-0 translate-y-4"
          }`}
        >
          <h3 className="text-xl font-semibold mb-6">Start from a template</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {templates.map((template, index) => (
              <Link key={template.id} href="/editor">
                <Card
                  className={`group cursor-pointer hover:border-accent hover:shadow-lg hover:shadow-accent/30 transition-all duration-300 p-6 text-center scale-in ${
                    animateIn ? "opacity-100 translate-y-0" : "opacity-0 translate-y-4"
                  }`}
                  style={{
                    transitionDelay: animateIn ? `${200 + index * 50}ms` : "0ms",
                  }}
                >
                  <div className="text-4xl mb-3 group-hover:scale-110 transition-transform duration-300 float-animation">
                    {template.icon}
                  </div>
                  <p className="font-medium group-hover:text-accent transition-colors">{template.name}</p>
                </Card>
              </Link>
            ))}
          </div>
        </div>

        {/* Search and Filter */}
        <div className="mb-8">
          <div className="flex items-center gap-4">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" size={20} />
              <input
                type="text"
                placeholder="Search documents..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-10 pr-4 py-2 bg-muted border border-border rounded-lg text-sm outline-none focus:border-accent focus:shadow-lg focus:shadow-accent/30 transition-all"
              />
            </div>
            <Button variant="outline" className="hover:bg-accent/20 hover:text-accent bg-transparent">
              Filter
            </Button>
          </div>
        </div>

        {/* Recent Projects */}
        <div>
          <h3 className="text-xl font-semibold mb-6 flex items-center gap-2">
            <Clock size={20} />
            Recent Projects {filteredProjects.length > 0 && `(${filteredProjects.length})`}
          </h3>

          {filteredProjects.length > 0 ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {filteredProjects.map((project, index) => (
                <div
                  key={project.id}
                  className={`group transition-all duration-700 ${
                    deletingId === project.id ? "opacity-0 scale-95" : "opacity-100 scale-100"
                  } ${animateIn ? "opacity-100 translate-y-0" : "opacity-0 translate-y-4"}`}
                  style={{
                    transitionDelay: animateIn ? `${400 + index * 50}ms` : "0ms",
                  }}
                >
                  <Link href="/editor">
                    <Card className="cursor-pointer hover:border-accent hover:shadow-lg hover:shadow-accent/30 transition-all duration-300 overflow-hidden">
                      <div
                        className={`h-32 ${project.thumbnail} mb-4 group-hover:scale-105 transition-transform duration-300 shimmer`}
                      />
                      <div className="p-4">
                        <h4 className="font-semibold group-hover:text-accent transition-colors">{project.name}</h4>
                        <p className="text-sm text-muted-foreground">{project.date}</p>
                        <p className="text-xs text-muted-foreground mt-2">{project.size}</p>
                      </div>
                    </Card>
                  </Link>
                  <div className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity flex gap-2">
                    <button
                      onClick={() => handleShareProject(project.name)}
                      className="p-2 bg-primary/80 hover:bg-primary text-primary-foreground rounded-lg transition-all hover:scale-110 glow-effect"
                      title="Share"
                    >
                      <Share2 size={16} />
                    </button>
                    <button
                      onClick={() => handleDeleteProject(project.id)}
                      className="p-2 bg-destructive/80 hover:bg-destructive text-destructive-foreground rounded-lg transition-all hover:scale-110"
                      title="Delete"
                    >
                      <Trash2 size={16} />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <Card className="p-12 text-center">
              <FileText size={48} className="mx-auto mb-4 text-muted-foreground opacity-50" />
              <p className="text-muted-foreground">No documents found</p>
              <p className="text-sm text-muted-foreground">Try adjusting your search</p>
            </Card>
          )}
        </div>
      </main>
    </div>
  )
}
