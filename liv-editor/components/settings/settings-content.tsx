"use client"

import Link from "next/link"
import { Button } from "@/components/ui/button"
import { CollapsibleCard } from "@/components/ui/collapsible-card"
import { ArrowLeft, Moon, Sun, Bell, Lock, Users, Trash2 } from "lucide-react"
import { useState } from "react"

export function SettingsContent() {
  const [theme, setTheme] = useState<"light" | "dark">("dark")
  const [notifications, setNotifications] = useState(true)
  const [autoSave, setAutoSave] = useState(true)

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border bg-muted/50 backdrop-blur-sm sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
          <Link href="/dashboard">
            <Button variant="outline" size="sm" className="gap-2 bg-transparent">
              <ArrowLeft size={16} />
              Back
            </Button>
          </Link>
          <h1 className="text-xl font-semibold">Settings</h1>
          <div className="w-10" />
        </div>
      </header>

      <main className="max-w-2xl mx-auto px-6 py-12">
        <div className="space-y-6">
          {/* Appearance Section */}
          <CollapsibleCard title="Appearance" icon={<Sun size={18} />} defaultOpen>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium">Theme</p>
                  <p className="text-sm text-muted-foreground">Choose your preferred color scheme</p>
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => setTheme("light")}
                    className={`px-4 py-2 rounded-lg transition-colors ${
                      theme === "light"
                        ? "bg-primary text-primary-foreground"
                        : "bg-muted text-muted-foreground hover:bg-muted/80"
                    }`}
                  >
                    <Sun size={18} />
                  </button>
                  <button
                    onClick={() => setTheme("dark")}
                    className={`px-4 py-2 rounded-lg transition-colors ${
                      theme === "dark"
                        ? "bg-primary text-primary-foreground"
                        : "bg-muted text-muted-foreground hover:bg-muted/80"
                    }`}
                  >
                    <Moon size={18} />
                  </button>
                </div>
              </div>

              <div>
                <label className="text-sm font-medium">Accent Color</label>
                <div className="flex gap-2 mt-2">
                  {["#6366f1", "#3b82f6", "#8b5cf6", "#ec4899"].map((color) => (
                    <button
                      key={color}
                      className="w-8 h-8 rounded-lg border-2 border-border hover:border-accent transition-colors"
                      style={{ backgroundColor: color }}
                    />
                  ))}
                </div>
              </div>
            </div>
          </CollapsibleCard>

          {/* Editor Section */}
          <CollapsibleCard title="Editor" defaultOpen>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium">Auto-save</p>
                  <p className="text-sm text-muted-foreground">Automatically save your work</p>
                </div>
                <input
                  type="checkbox"
                  checked={autoSave}
                  onChange={(e) => setAutoSave(e.target.checked)}
                  className="w-5 h-5 cursor-pointer"
                />
              </div>

              <div>
                <label className="text-sm font-medium">Auto-save Interval</label>
                <select className="w-full mt-2 px-3 py-2 bg-background border border-border rounded-md text-sm outline-none focus:border-accent transition-colors">
                  <option>Every 30 seconds</option>
                  <option>Every 1 minute</option>
                  <option>Every 5 minutes</option>
                </select>
              </div>

              <div>
                <label className="text-sm font-medium">Default Export Format</label>
                <select className="w-full mt-2 px-3 py-2 bg-background border border-border rounded-md text-sm outline-none focus:border-accent transition-colors">
                  <option>PDF</option>
                  <option>PNG</option>
                  <option>SVG</option>
                  <option>HTML</option>
                </select>
              </div>
            </div>
          </CollapsibleCard>

          {/* Notifications Section */}
          <CollapsibleCard title="Notifications" icon={<Bell size={18} />}>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium">Email Notifications</p>
                  <p className="text-sm text-muted-foreground">Receive updates about your documents</p>
                </div>
                <input
                  type="checkbox"
                  checked={notifications}
                  onChange={(e) => setNotifications(e.target.checked)}
                  className="w-5 h-5 cursor-pointer"
                />
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium">Collaboration Alerts</p>
                  <p className="text-sm text-muted-foreground">Get notified when others edit your documents</p>
                </div>
                <input type="checkbox" defaultChecked className="w-5 h-5 cursor-pointer" />
              </div>
            </div>
          </CollapsibleCard>

          {/* Privacy & Security Section */}
          <CollapsibleCard title="Privacy & Security" icon={<Lock size={18} />}>
            <div className="space-y-4">
              <div>
                <label className="text-sm font-medium">Default Document Visibility</label>
                <select className="w-full mt-2 px-3 py-2 bg-background border border-border rounded-md text-sm outline-none focus:border-accent transition-colors">
                  <option>Private</option>
                  <option>Shared with team</option>
                  <option>Public</option>
                </select>
              </div>

              <Button variant="outline" className="w-full gap-2 bg-transparent">
                <Lock size={16} />
                Change Password
              </Button>
            </div>
          </CollapsibleCard>

          {/* Team Section */}
          <CollapsibleCard title="Team & Collaboration" icon={<Users size={18} />}>
            <div className="space-y-4">
              <div>
                <p className="font-medium mb-3">Team Members</p>
                <div className="space-y-2">
                  {[
                    { name: "You", email: "you@example.com", role: "Owner" },
                    { name: "John Doe", email: "john@example.com", role: "Editor" },
                  ].map((member) => (
                    <div key={member.email} className="flex items-center justify-between p-3 bg-muted rounded-lg">
                      <div>
                        <p className="font-medium text-sm">{member.name}</p>
                        <p className="text-xs text-muted-foreground">{member.email}</p>
                      </div>
                      <span className="text-xs font-medium text-muted-foreground">{member.role}</span>
                    </div>
                  ))}
                </div>
              </div>

              <Button variant="outline" className="w-full bg-transparent">
                Invite Team Member
              </Button>
            </div>
          </CollapsibleCard>

          {/* Danger Zone */}
          <CollapsibleCard title="Danger Zone">
            <div className="space-y-4">
              <Button
                variant="outline"
                className="w-full gap-2 bg-transparent text-destructive hover:bg-destructive/10"
              >
                <Trash2 size={16} />
                Delete Account
              </Button>
              <p className="text-xs text-muted-foreground">
                This action cannot be undone. All your documents will be permanently deleted.
              </p>
            </div>
          </CollapsibleCard>
        </div>
      </main>
    </div>
  )
}
