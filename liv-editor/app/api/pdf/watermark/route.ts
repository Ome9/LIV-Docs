import { NextRequest, NextResponse } from 'next/server'
import { spawn } from 'child_process'
import path from 'path'

export async function POST(request: NextRequest) {
  try {
    const { inputPath, outputPath, watermarkText, watermarkImage, opacity, position } = await request.json()

    if (!inputPath || !outputPath || (!watermarkText && !watermarkImage)) {
      return NextResponse.json(
        { success: false, error: 'Invalid input parameters' },
        { status: 400 }
      )
    }

    // Path to the liv-pdf binary
    const binaryPath = path.join(process.cwd(), '..', 'bin', 'liv-pdf.exe')
    
    return new Promise((resolve) => {
      const args = ['watermark', '-i', inputPath, '-o', outputPath]
      
      if (watermarkText) {
        args.push('-t', watermarkText)
      }
      
      if (watermarkImage) {
        args.push('-img', watermarkImage)
      }
      
      if (opacity) {
        args.push('-opacity', opacity.toString())
      }
      
      if (position) {
        args.push('-pos', position)
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
          resolve(NextResponse.json({ success: true, message: 'Watermark added successfully' }))
        } else {
          resolve(NextResponse.json(
            { success: false, error: stderr || 'Failed to add watermark' },
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