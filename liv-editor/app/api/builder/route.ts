import { NextRequest, NextResponse } from 'next/server'
import { spawn } from 'child_process'
import path from 'path'

export async function POST(request: NextRequest) {
  try {
    const { inputPath, outputPath, manifest } = await request.json()

    if (!inputPath || !outputPath) {
      return NextResponse.json(
        { success: false, error: 'Invalid input parameters' },
        { status: 400 }
      )
    }

    // Path to the liv-builder binary
    const binaryPath = path.join(process.cwd(), '..', 'bin', 'builder.exe')
    
    return new Promise((resolve) => {
      const args = ['-i', inputPath, '-o', outputPath]
      
      if (manifest) {
        args.push('-m', manifest)
      }
      
      const process = spawn(binaryPath, args)

      let stdout = ''
      let stderr = ''

      process.stdout.on('data', (data) => {
        stdout += data.toString()
      })

      process.stderr.on('data', (data) => {
        stderr += data.toString()
      })

      process.on('close', (code) => {
        if (code === 0) {
          resolve(NextResponse.json({ success: true, message: 'LIV document built successfully' }))
        } else {
          resolve(NextResponse.json(
            { success: false, error: stderr || 'Failed to build LIV document' },
            { status: 500 }
          ))
        }
      })

      process.on('error', (error) => {
        resolve(NextResponse.json(
          { success: false, error: error.message },
          { status: 500 }
        ))
      })
    })
  } catch (error) {
    return NextResponse.json(
      { success: false, error: 'Internal server error' },
      { status: 500 }
    )
  }
}