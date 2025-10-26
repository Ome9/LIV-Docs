import { NextRequest, NextResponse } from 'next/server'

// This route is a placeholder - file dialogs are handled client-side in Electron
export async function POST(request: NextRequest) {
  return NextResponse.json(
    { success: false, error: 'File dialogs must be called from client-side Electron context' },
    { status: 400 }
  )
}